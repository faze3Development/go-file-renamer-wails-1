<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import toast from "svelte-french-toast";
  import { ProcessBulkFiles, GetBulkProcessingJob } from "../../../wailsjs/go/advanced_file_operations/AdvancedFileOperations.js";
  import { toErrorPayload, type ErrorPayload, type ErrorSeverity } from "../../lib/errorBus";
  import { advanced_file_operations } from "../../../wailsjs/go/models";
  import Toggle from "../shared/Toggle.svelte";
  import Button from "../shared/Button.svelte";

  interface Config {
    NamePattern?: string;
    NamerID?: string;
    [key: string]: any;
  }

  interface SelectedFile {
    name: string;
    size: number;
    type: string;
    base64: string;
  }

  interface BulkOptions {
    renameFiles: boolean;
    removeMetadata: boolean;
    optimizeFiles: boolean;
    compressFiles: boolean;
    pattern: string;
    namer: string;
    renameOptions: {
      preserveOriginalName: boolean;
      addTimestamp: boolean;
      addRandomId: boolean;
      addCustomDate: boolean;
      customDate: string;
      useRegexReplace: boolean;
      regexFind: string;
      regexReplace: string;
      sequentialNaming: {
        enabled: boolean;
        baseName: string;
        startIndex: number;
        padLength: number;
        keepExtension: boolean;
      };
    };
    allowedTypes: string[];
    maxFileSize: number;
  }

  interface ProcessingResult {
    filename: string;
    newName?: string;
    success: boolean;
    error?: string;
    action?: string;
    contentType?: string;
    contentBase64?: string;
  }

  interface BulkProcessingResponse {
    jobId?: string;
    status?: string;
    results?: ProcessingResult[];
    successCount?: number;
    failureCount?: number;
    durationMs?: number;
    totalFiles?: number;
    files?: any[];
  }

  export let config: Config;

  const DEFAULT_ALLOWED_TYPES = [
    "image/jpeg",
    "image/png",
    "image/gif",
    "application/pdf",
  ];

  const MAX_FILES_PER_BATCH = 100;

  let bulkOptions: BulkOptions = {
    renameFiles: true,
    removeMetadata: true,
    optimizeFiles: false,
    compressFiles: false,
    pattern: config?.NamePattern || "",
    namer: config?.NamerID || "template",
    renameOptions: {
      preserveOriginalName: false,
      addTimestamp: true,
      addRandomId: false,
      addCustomDate: false,
      customDate: "",
      useRegexReplace: false,
      regexFind: "",
      regexReplace: "",
      sequentialNaming: {
        enabled: false,
        baseName: "IMG",
        startIndex: 1,
        padLength: 3,
        keepExtension: true,
      },
    },
    allowedTypes: [...DEFAULT_ALLOWED_TYPES],
    maxFileSize: 50 * 1024 * 1024,
  };

  let allowedTypesInput = bulkOptions.allowedTypes.join(", ");
  let maxFileSizeInput = Math.round(bulkOptions.maxFileSize / (1024 * 1024));

  let selectedFiles: SelectedFile[] = [];
  let bulkResults: BulkProcessingResponse | null = null;
  let jobId: string = "";
  let jobStatus: string = "";
  let resultError: string = "";
  let isProcessing = false;
  let isRefreshing = false;
  let lastUpdated = null;
  let optionsInitialised = false;
  let patternTouched = false;
  let validationAttempted = false;
  let validationErrors: string[] = [];
  let patternError: string | null = null;

  const dispatch = createEventDispatcher<{
    error: ErrorPayload;
  }>();

  function notifyError(error: unknown, context: string, severity: ErrorSeverity = "error"): void {
    dispatch("error", toErrorPayload(error, context, severity));
  }

  $: totalSelectedSize = selectedFiles.reduce((sum, file) => sum + file.size, 0);
  $: allowedTypesInput = (bulkOptions.allowedTypes || []).join(", ");

  $: if (!optionsInitialised && config) {
    const pattern = config.NamePattern || bulkOptions.pattern || "";
    const namer = config.NamerID || bulkOptions.namer || "template";
    bulkOptions = { ...bulkOptions, pattern, namer };
    optionsInitialised = true;
  }

  $: if (optionsInitialised && config && selectedFiles.length === 0) {
    const updates: Partial<BulkOptions> = {};
    if (config.NamePattern && config.NamePattern !== bulkOptions.pattern) {
      updates.pattern = config.NamePattern;
    }
    if (config.NamerID && config.NamerID !== bulkOptions.namer) {
      updates.namer = config.NamerID;
    }
    if (Object.keys(updates).length) {
      bulkOptions = { ...bulkOptions, ...updates };
    }
  }

  function parseAllowedTypes(value: string): string[] {
    return value
      .split(/[,\n]+/)
      .map((entry) => entry.trim())
      .filter(Boolean);
  }


  function formatBytes(bytes: number): string {
    if (!bytes && bytes !== 0) return "-";
    const thresh = 1024;
    if (Math.abs(bytes) < thresh) {
      return `${bytes} B`;
    }
    const units = ["KB", "MB", "GB"];
    let u = -1;
    do {
      bytes /= thresh;
      ++u;
    } while (Math.abs(bytes) >= thresh && u < units.length - 1);
    return `${bytes.toFixed(1)} ${units[u]}`;
  }

  function removeFile(index: number): void {
    clearValidationErrors();
    selectedFiles = [
      ...selectedFiles.slice(0, index),
      ...selectedFiles.slice(index + 1),
    ];
  }

  function clearFiles() {
    clearValidationErrors();
    selectedFiles = [];
    bulkResults = null;
    jobId = "";
    jobStatus = "";
    lastUpdated = null;
    resultError = "";
  }

  function clearValidationErrors() {
    validationAttempted = false;
    validationErrors = [];
  }

  function onOptionToggle(key, value) {
    clearValidationErrors();
    bulkOptions = { ...bulkOptions, [key]: value };
    if (key === "renameFiles" && value === false) {
      patternTouched = false;
    }
  }

  function updateRenameOptions(patch) {
    clearValidationErrors();
    bulkOptions = {
      ...bulkOptions,
      renameOptions: {
        ...bulkOptions.renameOptions,
        ...patch,
      },
    };
  }

  function updateSequentialOptions(patch) {
    clearValidationErrors();
    updateRenameOptions({
      sequentialNaming: {
        ...bulkOptions.renameOptions.sequentialNaming,
        ...patch,
      },
    });
  }

  function onMaxFileSizeBlur() {
    clearValidationErrors();
    const value = Number(maxFileSizeInput);
    const normalised = Number.isFinite(value) && value > 0 ? value : 50;
    maxFileSizeInput = normalised;
    bulkOptions = {
      ...bulkOptions,
      maxFileSize: normalised * 1024 * 1024,
    };
  }

  async function ingestFiles(fileList: FileList | File[] | null): Promise<void> {
    clearValidationErrors();
    const files = Array.from(fileList || []);
    if (!files.length) {
      return;
    }

    if (selectedFiles.length + files.length > MAX_FILES_PER_BATCH) {
      const message = `A maximum of ${MAX_FILES_PER_BATCH} files per batch is supported.`;
      toast.error(message);
      notifyError(message, "file-ingest");
      return;
    }

    const newEntries: SelectedFile[] = [];
    for (const file of files) {
      try {
        const base64 = await readFileAsBase64(file);
        newEntries.push({
          name: file.name,
          size: file.size,
          type: file.type,
          base64,
        });
      } catch (err) {
        console.error("Failed to read file", file.name, err);
        const errorMessage = err instanceof Error ? err.message : String(err);
        const message = `Could not read ${file.name || "file"}: ${errorMessage}`;
        toast.error(message);
        notifyError(err, "file-ingest");
      }
    }

    if (newEntries.length) {
      selectedFiles = [...selectedFiles, ...newEntries];
    }
  }

  async function handleFileSelection(event) {
    await ingestFiles(event.currentTarget.files);
    event.target.value = "";
  }

  async function handleFileDrop(event) {
    try {
      event.stopPropagation?.();
      const droppedFiles = event.dataTransfer?.files;
      if (!droppedFiles?.length) {
        return;
      }
      await ingestFiles(droppedFiles);
    } catch (err) {
      console.error("Failed to process dropped files", err);
      const message = `Could not add dropped files: ${err?.message || err}`;
      toast.error(message);
      notifyError(err, "file-drop");
    }
  }

  function readFileAsBase64(file: File): Promise<string> {
    return new Promise((resolve, reject) => {
      const reader = new FileReader();
      reader.onerror = () => reject(reader.error || new Error("Unknown file read error"));
      reader.onload = () => {
        const result = reader.result;
        if (typeof result === "string") {
          const [, base64 = ""] = result.split(",");
          resolve(base64);
        } else {
          reject(new Error("Unexpected file reader result"));
        }
      };
      reader.readAsDataURL(file);
    });
  }

  function validateFilesBeforeProcessing() {
    if (!selectedFiles.length) {
      toast("Add at least one file to process.");
      return false;
    }

    const oversize = selectedFiles.filter((file) => file.size > bulkOptions.maxFileSize);
    if (oversize.length) {
      const message = `${oversize.length} file(s) exceed the max size of ${formatBytes(bulkOptions.maxFileSize)}.`;
      toast.error(message);
      notifyError(message, "file-validation");
      return false;
    }

    return true;
  }

  function normalizePatternInput(pattern: string): string {
    const trimmed = (pattern || "").trim();
    if (!trimmed) {
      return "";
    }

    const lowered = trimmed.toLowerCase();
    if (lowered === "default" || lowered === "random" || lowered === "timestamp") {
      return lowered;
    }

    if (trimmed.startsWith("seq:")) {
      return `seq:${trimmed.slice(4)}`;
    }
    if (trimmed.startsWith("SEQ:")) {
      return `seq:${trimmed.slice(4)}`;
    }

    if (trimmed.startsWith("regex:")) {
      return `regex:${trimmed.slice(6)}`;
    }
    if (trimmed.startsWith("REGEX:")) {
      return `regex:${trimmed.slice(6)}`;
    }

    if (trimmed.startsWith("date:")) {
      return trimmed;
    }
    if (trimmed.startsWith("DATE:")) {
      return `date:${trimmed.slice(5)}`;
    }

    return trimmed;
  }

  function validatePattern(pattern: string, options?: { preserveOriginalName?: boolean }): string | null {
    const trimmed = (pattern || "").trim();
    if (!trimmed || trimmed === "default") {
      return null;
    }

    const segments = trimmed
      .split("|")
      .map((segment) => segment.trim())
      .filter(Boolean);
    if (segments.length > 1) {
      for (const segment of segments) {
        const segmentError = validatePattern(segment, options);
        if (segmentError) {
          return segmentError;
        }
      }
      return null;
    }

    const lowered = trimmed.toLowerCase();
    if (lowered === "random" || lowered === "timestamp") {
      return null;
    }

    if (lowered.startsWith("date:")) {
      return null;
    }

    if (lowered.startsWith("seq:")) {
      const parts = trimmed.split(":");
      const maybeStart = parts[2];
      const maybePad = parts[3];
      const maybeKeep = parts[4];
      if (maybeStart && Number.isNaN(Number(maybeStart))) {
        return "Sequential pattern start index must be a number.";
      }
      if (maybePad && Number.isNaN(Number(maybePad))) {
        return "Sequential pattern pad length must be a number.";
      }
      if (maybeKeep && !["0", "1", "true", "false"].includes(maybeKeep.toLowerCase())) {
        return "Sequential pattern keep-extension flag must be 0, 1, true, or false.";
      }
      if ((parts[1] ?? "").trim().length === 0 && !options?.preserveOriginalName) {
        return "Provide a base name or enable preserve original name for sequential naming.";
      }
      return null;
    }

    if (lowered.startsWith("regex:")) {
      const body = trimmed.slice(6);
      const separator = body.indexOf(":");
      if (separator === -1) {
        return "Regex patterns must use the format regex:find:replace.";
      }
      const patternBody = body.slice(0, separator);
      try {
        new RegExp(patternBody);
      } catch (err) {
        const errorMessage = err instanceof Error ? err.message : String(err);
        return `Regex pattern is invalid: ${errorMessage}`;
      }
      return null;
    }

    return "Unsupported rename pattern. Use random, timestamp, seq:, regex:, date:, or leave blank.";
  }

  function getPatternError(value: string): string | null {
    const normalized = normalizePatternInput(value);
    return validatePattern(normalized, {
      preserveOriginalName: bulkOptions.renameOptions.preserveOriginalName,
    });
  }

  function validateConfiguration(candidateOptions: BulkOptions): string[] {
    const errors: string[] = [];

    if (!candidateOptions.renameFiles && !candidateOptions.removeMetadata && !candidateOptions.optimizeFiles && !candidateOptions.compressFiles) {
      errors.push("Enable at least one processing option (rename, metadata removal, optimization, or compression).");
    }

    if (candidateOptions.renameFiles) {
      const patternIssue = validatePattern(candidateOptions.pattern || "", {
        preserveOriginalName: candidateOptions.renameOptions?.preserveOriginalName,
      });
      if (patternIssue) {
        errors.push(patternIssue);
      }

      if (candidateOptions.renameOptions?.sequentialNaming?.enabled) {
        const base = candidateOptions.renameOptions.sequentialNaming.baseName?.trim();
        if (!base && !candidateOptions.renameOptions.preserveOriginalName) {
          errors.push("When sequential naming is enabled, provide a base name or enable preserve original name.");
        }
      }
    }

    const allowedTypes = candidateOptions.allowedTypes ?? [];
    if (!allowedTypes.length) {
      errors.push("Allowed content types list cannot be empty.");
    }

    return errors;
  }

  $: patternError = bulkOptions.renameFiles ? getPatternError(bulkOptions.pattern) : null;

  function buildRequestPayload() {
    // Build properly typed request using Wails-generated models
    const request = new advanced_file_operations.BulkProcessingRequest({
      files: selectedFiles.map((file) => new advanced_file_operations.BulkProcessingFile({
        filename: file.name,
        contentBase64: file.base64,
        contentType: file.type || "",
        size: file.size,
      })),
      options: new advanced_file_operations.BulkProcessingOptions({
        renameFiles: bulkOptions.renameFiles,
        removeMetadata: bulkOptions.removeMetadata,
        optimizeFiles: bulkOptions.optimizeFiles,
        compressFiles: bulkOptions.compressFiles,
        pattern: bulkOptions.pattern,
        namer: bulkOptions.namer,
        renameOptions: bulkOptions.renameOptions,
        allowedTypes:
          bulkOptions.allowedTypes && bulkOptions.allowedTypes.length
            ? bulkOptions.allowedTypes
            : [...DEFAULT_ALLOWED_TYPES],
        maxFileSize: bulkOptions.maxFileSize,
      }),
    });

    return request;
  }

  async function processSelectedFiles() {
    if (!validateFilesBeforeProcessing()) {
      return;
    }

    const normalizedPattern = normalizePatternInput(bulkOptions.pattern);
    const trimmedSequentialBase = bulkOptions.renameOptions.sequentialNaming.baseName?.trim() ?? "";
    const candidateOptions: BulkOptions = {
      ...bulkOptions,
      pattern: normalizedPattern,
      renameOptions: {
        ...bulkOptions.renameOptions,
        sequentialNaming: {
          ...bulkOptions.renameOptions.sequentialNaming,
          baseName: trimmedSequentialBase,
        },
      },
    };

    validationAttempted = true;
    const configurationErrors = validateConfiguration(candidateOptions);
    validationErrors = configurationErrors;
    if (configurationErrors.length) {
      patternTouched = true;
      toast.error(configurationErrors[0]);
      return;
    }

    if (candidateOptions.pattern !== bulkOptions.pattern || trimmedSequentialBase !== bulkOptions.renameOptions.sequentialNaming.baseName) {
      bulkOptions = candidateOptions;
    }

    validationAttempted = false;
    validationErrors = [];

    isProcessing = true;
    resultError = "";
    bulkResults = null;

    try {
      const response = await ProcessBulkFiles(buildRequestPayload());
      bulkResults = response;
      jobId = response?.jobId || "";
      jobStatus = response ? (response.failureCount > 0 ? "completed_with_errors" : "completed") : "";
      lastUpdated = new Date();
      toast.success("Advanced processing complete.");
    } catch (err) {
      console.error("Bulk processing failed", err);
      resultError = err?.message || String(err);
      toast.error(`Processing failed: ${resultError}`);
      notifyError(err, "bulk-processing");
    } finally {
      isProcessing = false;
    }
  }

  async function refreshJobStatus() {
    if (!jobId) {
      return;
    }

    isRefreshing = true;
    resultError = "";

    try {
      const job = await GetBulkProcessingJob(jobId);
      jobStatus = job?.status || jobStatus;
      lastUpdated = new Date();

      if (job) {
        const existingByFilename = new Map(
          (bulkResults?.results || []).map((result) => [result.filename, result]),
        );

        bulkResults = {
          jobId: job.id,
          totalFiles: job.files?.length ?? 0,
          successCount: job.results?.filter((item) => item.success).length ?? 0,
          failureCount: job.results?.filter((item) => !item.success).length ?? 0,
          durationMs: job.durationMs ?? bulkResults?.durationMs ?? 0,
          results:
            job.results?.map((item) => {
              const existing = existingByFilename.get(item.filename);
              return {
                filename: item.filename,
                newName: item.newName,
                success: item.success,
                error: item.error,
                action: item.action,
                contentType: item.contentType,
                contentBase64: existing?.contentBase64,
              };
            }) ?? [],
        };
      }
    } catch (err) {
      console.error("Failed to refresh job state", err);
      resultError = err?.message || String(err);
      toast.error(`Could not refresh job: ${resultError}`);
      notifyError(err, "job-refresh");
    } finally {
      isRefreshing = false;
    }
  }

  function downloadProcessedFile(result: ProcessingResult): void {
    if (!result?.contentBase64) {
      toast("Run a new batch to retrieve downloads for this job.");
      return;
    }

    const link = document.createElement("a");
    link.href = `data:${result.contentType || "application/octet-stream"};base64,${result.contentBase64}`;
    link.download = result.newName || result.filename || "processed-file";
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  }

  function formatDuration(durationMs) {
    if (!durationMs && durationMs !== 0) return "-";
    if (durationMs < 1000) return `${durationMs} ms`;
    const seconds = durationMs / 1000;
    if (seconds < 60) return `${seconds.toFixed(2)} s`;
    const minutes = Math.floor(seconds / 60);
    const rem = seconds % 60;
    return `${minutes}m ${rem.toFixed(0)}s`;
  }

  function isAllowedType(type) {
    if (!bulkOptions.allowedTypes?.length || !type) {
      return true;
    }
    return bulkOptions.allowedTypes.includes(type);
  }
</script>

<div class="config-card">
  <div class="card-header">
    <h3>🧰 Bulk Processing Toolkit</h3>
    <p>Run ad-hoc files through the advanced operations pipeline.</p>
  </div>
  <div class="card-content bulk-operations">
    <div class="file-picker">
      <label
        class="upload-area"
        on:dragover|preventDefault
        on:dragenter|preventDefault
        on:drop|preventDefault={handleFileDrop}
      >
        <input type="file" multiple on:change={handleFileSelection} />
        <span class="upload-icon">📤</span>
        <span class="upload-text">
          Drop files here or <strong>click to browse</strong>
          <small>Up to {MAX_FILES_PER_BATCH} files per batch</small>
        </span>
      </label>
      {#if selectedFiles.length}
        <div class="file-summary">
          <span>{selectedFiles.length} files selected • {formatBytes(totalSelectedSize)}</span>
          <button class="link-btn" on:click={clearFiles}>Clear selection</button>
        </div>
        <ul class="file-list">
          {#each selectedFiles as file, index}
            <li class:disallowed={!isAllowedType(file.type)}>
              <div class="file-meta">
                <strong>{file.name}</strong>
                <small>{file.type || "Unknown"} • {formatBytes(file.size)}</small>
                {#if !isAllowedType(file.type)}
                  <span class="badge warn">Type not in allowed list</span>
                {/if}
              </div>
              <button class="link-btn" on:click={() => removeFile(index)}>Remove</button>
            </li>
          {/each}
        </ul>
      {/if}
    </div>

    <div class="options-grid">
      <div class="option-toggle">
        <Toggle
          checked={bulkOptions.renameFiles}
          on:change={(e) => onOptionToggle("renameFiles", e.detail.checked)}
          ariaLabel="Rename Files"
        />
        <span class="toggle-label">
          <strong>Rename Files</strong>
          <small>Apply naming templates or sequential patterns</small>
        </span>
      </div>

      <div class="option-toggle">
        <Toggle
          checked={bulkOptions.removeMetadata}
          on:change={(e) => onOptionToggle("removeMetadata", e.detail.checked)}
          ariaLabel="Remove Metadata"
        />
        <span class="toggle-label">
          <strong>Remove Metadata</strong>
          <small>Strip EXIF and embedded metadata (requires ExifTool)</small>
        </span>
      </div>

      <div class="option-toggle">
        <Toggle
          checked={bulkOptions.optimizeFiles}
          on:change={(e) => onOptionToggle("optimizeFiles", e.detail.checked)}
          ariaLabel="Optimize Files"
        />
        <span class="toggle-label">
          <strong>Optimize Files</strong>
          <small>Run format-specific optimisation for supported types</small>
        </span>
      </div>

      <div class="option-toggle">
        <Toggle
          checked={bulkOptions.compressFiles}
          on:change={(e) => onOptionToggle("compressFiles", e.detail.checked)}
          ariaLabel="Compress Output"
        />
        <span class="toggle-label">
          <strong>Compress Output</strong>
          <small>Apply lossless compression to processed files</small>
        </span>
      </div>
    </div>

    {#if bulkOptions.renameFiles}
      <div class="rename-settings">
        <h4>Rename Settings</h4>
        <div class="rename-grid">
          <label>
            <span>Pattern</span>
            <input
              type="text"
              bind:value={bulkOptions.pattern}
              placeholder={"Example: random | timestamp | seq:IMG:1:3"}
              class:input-error={patternTouched && !!patternError}
              on:input={clearValidationErrors}
              on:blur={() => (patternTouched = true)}
            />
            <small class="muted">Leave blank to keep names, or use presets like random, timestamp, seq:Base:Start:Pad, regex:find:replace.</small>
            {#if patternTouched && patternError}
              <small class="field-error">{patternError}</small>
            {/if}
          </label>
          <label>
            <span>Namer ID</span>
            <input
              type="text"
              bind:value={bulkOptions.namer}
              placeholder="random | template | sequential"
            />
          </label>
          <label class="inline-option">
            <input
              type="checkbox"
              checked={bulkOptions.renameOptions.preserveOriginalName}
              on:change={(e) =>
                updateRenameOptions({ preserveOriginalName: e.currentTarget.checked })}
            />
            <span>Preserve original base name</span>
          </label>
          <label class="inline-option">
            <input
              type="checkbox"
              checked={bulkOptions.renameOptions.addTimestamp}
              on:change={(e) =>
                updateRenameOptions({ addTimestamp: e.currentTarget.checked })}
            />
            <span>Append timestamp</span>
          </label>
          <label class="inline-option">
            <input
              type="checkbox"
              checked={bulkOptions.renameOptions.addRandomId}
              on:change={(e) =>
                updateRenameOptions({ addRandomId: e.currentTarget.checked })}
            />
            <span>Append random suffix</span>
          </label>
        </div>
        <div class="sequential-card">
          <div class="sequential-header">
            <label class="inline-option">
              <input
                type="checkbox"
                checked={bulkOptions.renameOptions.sequentialNaming.enabled}
                on:change={(e) =>
                  updateSequentialOptions({ enabled: e.currentTarget.checked })}
              />
              <span>Enable sequential naming</span>
            </label>
          </div>
          {#if bulkOptions.renameOptions.sequentialNaming.enabled}
            <div class="sequential-grid">
              <label>
                <span>Base name</span>
                <input
                  type="text"
                  bind:value={bulkOptions.renameOptions.sequentialNaming.baseName}
                  on:input={clearValidationErrors}
                />
              </label>
              <label>
                <span>Start index</span>
                <input
                  type="number"
                  min="0"
                  bind:value={bulkOptions.renameOptions.sequentialNaming.startIndex}
                  on:input={clearValidationErrors}
                />
              </label>
              <label>
                <span>Pad length</span>
                <input
                  type="number"
                  min="1"
                  bind:value={bulkOptions.renameOptions.sequentialNaming.padLength}
                  on:input={clearValidationErrors}
                />
              </label>
              <label class="inline-option">
                <input
                  type="checkbox"
                  checked={bulkOptions.renameOptions.sequentialNaming.keepExtension}
                  on:change={(e) =>
                    updateSequentialOptions({ keepExtension: e.currentTarget.checked })}
                />
                <span>Keep original extension</span>
              </label>
            </div>
          {/if}
        </div>
      </div>
    {/if}

    <div class="constraints-grid">
      <div class="info-field">
        <span class="field-label">Allowed content types</span>
        <!-- Present allowed types as non-editable informative text so users don't accidentally change them here -->
        <div class="allowed-types" aria-live="polite">{allowedTypesInput}</div>
        <small class="muted">Configured MIME types (comma-separated)</small>
      </div>
      <label>
        <span>Max file size (MB)</span>
        <input
          type="number"
          min="1"
          max="200"
          value={maxFileSizeInput}
          on:input={(e) => (maxFileSizeInput = Number(e.currentTarget.value))}
          on:blur={onMaxFileSizeBlur}
        />
        <small>Files larger than this limit are skipped</small>
      </label>
    </div>

    {#if validationErrors.length}
      <div class="validation-errors" role="alert">
        {#each validationErrors as validationError}
          <div>{validationError}</div>
        {/each}
      </div>
    {/if}

    <div class="actions">
      <Button
        variant="primary"
        on:click={processSelectedFiles}
        disabled={isProcessing || !selectedFiles.length}
      >
        {#if isProcessing}
          <span class="spinner" />
          Processing…
        {:else}
          Run Advanced Operations
        {/if}
      </Button>

      <Button
        variant="ghost"
        on:click={refreshJobStatus}
        disabled={!jobId || isRefreshing}
      >
        {#if isRefreshing}
          <span class="spinner" />
          Refreshing…
        {:else}
          Refresh last job
        {/if}
      </Button>
    </div>

    {#if resultError}
      <div class="result-error">
        ⚠️ {resultError}
      </div>
    {/if}

    {#if bulkResults}
      <div class="results-panel">
        <div class="results-header">
          <div>
            <strong>Job ID:</strong> {bulkResults.jobId}
          </div>
          <div class="results-meta">
            <span>Status: {jobStatus || "completed"}</span>
            <span>
              Success: {bulkResults.successCount} • Failures: {bulkResults.failureCount}
            </span>
            <span>Duration: {formatDuration(bulkResults.durationMs)}</span>
            {#if lastUpdated}
              <span>Updated: {lastUpdated.toLocaleTimeString()}</span>
            {/if}
          </div>
        </div>
        <table class="results-table">
          <thead>
            <tr>
              <th>File</th>
              <th>Action</th>
              <th>Status</th>
              <th class="min">Download</th>
            </tr>
          </thead>
          <tbody>
            {#each bulkResults.results as result}
              <tr class:has-error={!result.success}>
                <td>
                  <div class="result-name">
                    <strong>{result.newName || result.filename}</strong>
                    <small>{result.filename}</small>
                  </div>
                </td>
                <td>{result.action || "—"}</td>
                <td>
                  {#if result.success}
                    <span class="badge success">Success</span>
                  {:else}
                    <span class="badge warn">{result.error || "Failed"}</span>
                  {/if}
                </td>
                <td class="min">
                  {#if result.success}
                    <button
                      class="link-btn"
                      on:click={() => downloadProcessedFile(result)}
                    >
                      Download
                    </button>
                  {:else}
                    —
                  {/if}
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {/if}
  </div>
</div>

<style>
  .bulk-operations {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .file-picker {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .upload-area {
    border: 2px dashed rgba(139, 92, 246, 0.4);
    border-radius: 12px;
    padding: 1.25rem;
    display: flex;
    align-items: center;
    gap: 0.75rem;
    justify-content: center;
    flex-direction: column;
    color: var(--text-secondary);
    cursor: pointer;
    transition: border-color 0.2s ease, background 0.2s ease;
    background: rgba(139, 92, 246, 0.06);
    text-align: center;
  }

  .upload-area:hover {
    border-color: var(--accent-primary);
    background: rgba(139, 92, 246, 0.12);
  }

  .upload-area input[type="file"] {
    display: none;
  }

  .upload-icon {
    font-size: 2rem;
  }

  .upload-text strong {
    color: var(--text-primary);
  }

  .upload-text small {
    display: block;
    margin-top: 0.4rem;
    font-size: 0.75rem;
    color: var(--text-muted);
  }

  .file-summary {
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-size: 0.9rem;
    color: var(--text-secondary);
  }

  .file-list {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    max-height: 200px;
    overflow-y: auto;
    padding-right: 0.5rem;
  }

  .file-list li {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.5rem 0.75rem;
    border-radius: 8px;
    border: 1px solid rgba(255, 255, 255, 0.08);
    background: rgba(255, 255, 255, 0.03);
  }

  .file-list li.disallowed {
    border-color: rgba(239, 68, 68, 0.4);
  }

  .file-meta {
    display: flex;
    flex-direction: column;
    gap: 0.2rem;
  }

  .file-meta small {
    color: var(--text-muted);
    font-size: 0.7rem;
  }

  .options-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
    gap: 1rem;
  }

  .option-toggle {
    position: relative;
    display: flex;
    align-items: center;
    gap: 0.9rem;
    padding: 0.85rem 1rem;
    border-radius: 12px;
    border: 1px solid rgba(255, 255, 255, 0.08);
    background: rgba(0, 0, 0, 0.25);
    cursor: pointer;
    transition: border-color 0.2s ease, background 0.2s ease;
    overflow: visible; /* ensure toggle visuals aren't clipped */
  }

  .option-toggle:hover {
    border-color: rgba(139, 92, 246, 0.4);
    background: rgba(139, 92, 246, 0.08);
  }

  /* Toggle visuals are provided by the shared Toggle component; local slider rules removed. */
  .toggle-label {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .toggle-label strong {
    color: var(--text-primary);
    font-size: 0.95rem;
  }

  .toggle-label small {
    color: var(--text-muted);
    font-size: 0.75rem;
  }

  /* Toggle visuals are provided by the shared Toggle component; local slider rules removed. */
  

  .constraints-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
    gap: 1rem;
  }

  .constraints-grid small {
    font-size: 0.75rem;
    color: var(--text-muted);
  }

  .inline-option {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 0.85rem;
    color: var(--text-secondary);
  }

  .sequential-card {
    border: 1px dashed rgba(255, 255, 255, 0.15);
    border-radius: 10px;
    padding: 0.75rem;
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    background: rgba(0, 0, 0, 0.25);
  }

  .sequential-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
    gap: 0.75rem;
  }

  .actions {
    display: flex;
    gap: 0.75rem;
    flex-wrap: wrap;
    align-items: center;
    /* ensure buttons will shrink before overflowing the container */
    justify-content: flex-start;
  }

  .link-btn {
    border: none;
    border-radius: 10px;
    padding: 0.65rem 1.2rem;
    cursor: pointer;
    font-weight: 500;
    transition: all 0.2s ease;
  }

  /* Buttons are provided by shared Button component; keep link-btn for inline links */

  /* Informational field for allowed types */
  .info-field .allowed-types {
    padding: 0.6rem 0.75rem;
    border-radius: 8px;
    border: 1px solid rgba(255, 255, 255, 0.06);
    background: rgba(0, 0, 0, 0.35);
    color: var(--text-primary);
    font-size: 0.85rem;
    word-break: break-word;
    white-space: pre-wrap;
  }

  /* Make constraints grid responsive so controls wrap instead of overflowing */
  .constraints-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
    gap: 1rem;
    align-items: start;
  }

  .link-btn:hover:not(:disabled) {
    text-decoration: underline;
  }

  .badge {
    padding: 0.2rem 0.5rem;
    border-radius: 999px;
    font-size: 0.7rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.04em;
  }

  .badge.success {
    background: rgba(16, 185, 129, 0.2);
    color: #34d399;
  }

  .badge.warn {
    background: rgba(239, 68, 68, 0.15);
    color: #f87171;
  }

  .results-panel {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    max-height: 400px;
    overflow-y: auto;
  }

  .results-header {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    color: var(--text-secondary);
    flex-shrink: 0;
  }

  .results-header strong {
    color: var(--text-primary);
  }

  .results-meta {
    display: flex;
    flex-wrap: wrap;
    gap: 0.75rem;
    font-size: 0.8rem;
  }

  .results-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.85rem;
  }

  .results-table th,
  .results-table td {
    padding: 0.6rem;
    text-align: left;
    border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  }

  .results-table tr:last-child td {
    border-bottom: none;
  }

  .results-table tr.has-error {
    background: rgba(239, 68, 68, 0.05);
  }

  .results-table th.min,
  .results-table td.min {
    width: 120px;
  }

  .result-name {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .result-name small {
    color: var(--text-muted);
    font-size: 0.75rem;
  }

  .result-error {
    padding: 0.75rem 1rem;
    background: rgba(239, 68, 68, 0.12);
    border: 1px solid rgba(239, 68, 68, 0.3);
    border-radius: 10px;
    color: #f87171;
  }

  .input-error {
    border-color: var(--status-error, #ef4444);
    box-shadow: 0 0 0 1px rgba(239, 68, 68, 0.24);
  }

  .field-error {
    color: var(--status-error, #ef4444);
    font-size: 0.75rem;
    margin-top: 0.25rem;
    display: block;
  }

  .validation-errors {
    background: rgba(239, 68, 68, 0.08);
    border: 1px solid rgba(239, 68, 68, 0.25);
    border-radius: 10px;
    padding: 0.75rem 1rem;
    color: var(--status-error, #ef4444);
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
  }

  .validation-errors div::before {
    content: "⚠";
    margin-right: 0.35rem;
  }

  .spinner {
    width: 1rem;
    height: 1rem;
    border-radius: 50%;
    border: 2px solid rgba(255, 255, 255, 0.2);
    border-top-color: white;
    animation: spin 0.8s linear infinite;
    display: inline-block;
  }

  /* removed unused nested spinner selectors; .spinner is styled above and can be used standalone */

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }

  @media (max-width: 768px) {
    .results-meta {
      flex-direction: column;
    }

    .options-grid,
    .rename-grid,
    .constraints-grid,
    .sequential-grid {
      grid-template-columns: 1fr;
    }
    .constraints-grid {
      grid-template-columns: 1fr;
    }
  }
</style>
