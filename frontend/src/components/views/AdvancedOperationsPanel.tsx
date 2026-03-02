import React, { useState, useEffect, useMemo } from 'react'
import toast from 'react-hot-toast'
import { ProcessBulkFiles, GetBulkProcessingJob } from '../../../wailsjs/go/advanced_file_operations/AdvancedFileOperations.js'
import { advanced_file_operations } from '../../../wailsjs/go/models'
import { toErrorPayload } from '../../lib/errorBus'
import type { ErrorPayload, ErrorSeverity } from '../../lib/errorBus'
import Toggle from '../shared/Toggle'
import Button from '../shared/Button'

interface Config {
  NamePattern?: string
  NamerID?: string
  [key: string]: unknown
}

interface SelectedFile {
  name: string
  size: number
  type: string
  base64: string
}

interface SequentialNaming {
  enabled: boolean
  baseName: string
  startIndex: number
  padLength: number
  keepExtension: boolean
}

interface RenameOptions {
  preserveOriginalName: boolean
  addTimestamp: boolean
  addRandomId: boolean
  addCustomDate: boolean
  customDate: string
  useRegexReplace: boolean
  regexFind: string
  regexReplace: string
  sequentialNaming: SequentialNaming
}

interface BulkOptions {
  renameFiles: boolean
  removeMetadata: boolean
  optimizeFiles: boolean
  compressFiles: boolean
  pattern: string
  namer: string
  renameOptions: RenameOptions
  allowedTypes: string[]
  maxFileSize: number
}

interface ProcessingResult {
  filename: string
  newName?: string
  success: boolean
  error?: string
  action?: string
  contentType?: string
  contentBase64?: string
}

interface BulkProcessingResponse {
  jobId?: string
  status?: string
  results?: ProcessingResult[]
  successCount?: number
  failureCount?: number
  durationMs?: number
  totalFiles?: number
}

interface Props {
  config: Config
  onError: (payload: ErrorPayload) => void
}

const DEFAULT_ALLOWED_TYPES = ['image/jpeg', 'image/png', 'image/gif', 'application/pdf']
const MAX_FILES_PER_BATCH = 100

function formatBytes(bytes: number): string {
  if (!bytes && bytes !== 0) return '-'
  const thresh = 1024
  if (Math.abs(bytes) < thresh) return `${bytes} B`
  const units = ['KB', 'MB', 'GB']
  let u = -1
  let b = bytes
  do { b /= thresh; ++u } while (Math.abs(b) >= thresh && u < units.length - 1)
  return `${b.toFixed(1)} ${units[u]}`
}

function formatDuration(durationMs: number | undefined): string {
  if (!durationMs && durationMs !== 0) return '-'
  if (durationMs < 1000) return `${durationMs} ms`
  const seconds = durationMs / 1000
  if (seconds < 60) return `${seconds.toFixed(2)} s`
  const minutes = Math.floor(seconds / 60)
  const rem = seconds % 60
  return `${minutes}m ${rem.toFixed(0)}s`
}

function normalizePatternInput(pattern: string): string {
  const trimmed = (pattern || '').trim()
  if (!trimmed) return ''
  const lowered = trimmed.toLowerCase()
  if (lowered === 'default' || lowered === 'random' || lowered === 'timestamp') return lowered
  if (trimmed.startsWith('seq:') || trimmed.startsWith('SEQ:')) return `seq:${trimmed.slice(4)}`
  if (trimmed.startsWith('regex:') || trimmed.startsWith('REGEX:')) return `regex:${trimmed.slice(6)}`
  if (trimmed.startsWith('date:') || trimmed.startsWith('DATE:')) return `date:${trimmed.slice(5)}`
  return trimmed
}

function validatePattern(pattern: string, opts?: { preserveOriginalName?: boolean }): string | null {
  const trimmed = (pattern || '').trim()
  if (!trimmed || trimmed === 'default') return null
  const segments = trimmed.split('|').map((s) => s.trim()).filter(Boolean)
  if (segments.length > 1) {
    for (const seg of segments) { const e = validatePattern(seg, opts); if (e) return e }
    return null
  }
  const lowered = trimmed.toLowerCase()
  if (lowered === 'random' || lowered === 'timestamp' || lowered.startsWith('date:')) return null
  if (lowered.startsWith('seq:')) {
    const parts = trimmed.split(':')
    if (parts[2] && Number.isNaN(Number(parts[2]))) return 'Sequential pattern start index must be a number.'
    if (parts[3] && Number.isNaN(Number(parts[3]))) return 'Sequential pattern pad length must be a number.'
    if (parts[4] && !['0', '1', 'true', 'false'].includes(parts[4].toLowerCase())) return 'Sequential pattern keep-extension flag must be 0, 1, true, or false.'
    if ((parts[1] ?? '').trim().length === 0 && !opts?.preserveOriginalName) return 'Provide a base name or enable preserve original name for sequential naming.'
    return null
  }
  if (lowered.startsWith('regex:')) {
    const body = trimmed.slice(6)
    const sep = body.indexOf(':')
    if (sep === -1) return 'Regex patterns must use the format regex:find:replace.'
    try { new RegExp(body.slice(0, sep)) } catch (e) { return `Regex pattern is invalid: ${e instanceof Error ? e.message : e}` }
    return null
  }
  return 'Unsupported rename pattern. Use random, timestamp, seq:, regex:, date:, or leave blank.'
}

function readFileAsBase64(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onerror = () => reject(reader.error || new Error('Unknown file read error'))
    reader.onload = () => {
      if (typeof reader.result === 'string') {
        const [, base64 = ''] = reader.result.split(',')
        resolve(base64)
      } else {
        reject(new Error('Unexpected reader result'))
      }
    }
    reader.readAsDataURL(file)
  })
}

export default function AdvancedOperationsPanel({ config, onError }: Props) {
  const notifyError = (error: unknown, context: string, severity: ErrorSeverity = 'error') => {
    onError(toErrorPayload(error, context, severity))
  }

  const [bulkOptions, setBulkOptions] = useState<BulkOptions>({
    renameFiles: true,
    removeMetadata: true,
    optimizeFiles: false,
    compressFiles: false,
    pattern: config?.NamePattern || '',
    namer: (config?.NamerID as string) || 'template',
    renameOptions: {
      preserveOriginalName: false,
      addTimestamp: true,
      addRandomId: false,
      addCustomDate: false,
      customDate: '',
      useRegexReplace: false,
      regexFind: '',
      regexReplace: '',
      sequentialNaming: { enabled: false, baseName: 'IMG', startIndex: 1, padLength: 3, keepExtension: true },
    },
    allowedTypes: [...DEFAULT_ALLOWED_TYPES],
    maxFileSize: 50 * 1024 * 1024,
  })
  const [maxFileSizeInput, setMaxFileSizeInput] = useState(50)
  const [selectedFiles, setSelectedFiles] = useState<SelectedFile[]>([])
  const [bulkResults, setBulkResults] = useState<BulkProcessingResponse | null>(null)
  const [jobId, setJobId] = useState('')
  const [jobStatus, setJobStatus] = useState('')
  const [resultError, setResultError] = useState('')
  const [isProcessing, setIsProcessing] = useState(false)
  const [isRefreshing, setIsRefreshing] = useState(false)
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null)
  const [patternTouched, setPatternTouched] = useState(false)
  const [validationErrors, setValidationErrors] = useState<string[]>([])

  // Sync config changes when no files are selected
  useEffect(() => {
    if (selectedFiles.length === 0) {
      setBulkOptions((prev) => ({
        ...prev,
        pattern: config?.NamePattern || prev.pattern,
        namer: (config?.NamerID as string) || prev.namer,
      }))
    }
  }, [config?.NamePattern, config?.NamerID, selectedFiles.length])

  const totalSelectedSize = useMemo(() => selectedFiles.reduce((s, f) => s + f.size, 0), [selectedFiles])
  const allowedTypesDisplay = bulkOptions.allowedTypes.join(', ')
  const patternError = bulkOptions.renameFiles
    ? validatePattern(bulkOptions.pattern, { preserveOriginalName: bulkOptions.renameOptions.preserveOriginalName })
    : null

  const clearValidation = () => setValidationErrors([])

  const updateRenameOptions = (patch: Partial<RenameOptions>) => {
    clearValidation()
    setBulkOptions((prev) => ({ ...prev, renameOptions: { ...prev.renameOptions, ...patch } }))
  }

  const updateSequentialOptions = (patch: Partial<SequentialNaming>) => {
    clearValidation()
    updateRenameOptions({ sequentialNaming: { ...bulkOptions.renameOptions.sequentialNaming, ...patch } })
  }

  const isAllowedType = (type: string) => {
    if (!bulkOptions.allowedTypes?.length || !type) return true
    return bulkOptions.allowedTypes.includes(type)
  }

  const ingestFiles = async (fileList: FileList | File[] | null) => {
    clearValidation()
    const files = Array.from(fileList || [])
    if (!files.length) return
    if (selectedFiles.length + files.length > MAX_FILES_PER_BATCH) {
      const message = `A maximum of ${MAX_FILES_PER_BATCH} files per batch is supported.`
      toast.error(message); notifyError(message, 'file-ingest'); return
    }
    const newEntries: SelectedFile[] = []
    for (const file of files) {
      try {
        const base64 = await readFileAsBase64(file)
        newEntries.push({ name: file.name, size: file.size, type: file.type, base64 })
      } catch (err) {
        const message = `Could not read ${file.name || 'file'}: ${err instanceof Error ? err.message : err}`
        toast.error(message); notifyError(err, 'file-ingest')
      }
    }
    if (newEntries.length) setSelectedFiles((prev) => [...prev, ...newEntries])
  }

  const handleFileSelection = async (e: React.ChangeEvent<HTMLInputElement>) => {
    await ingestFiles(e.currentTarget.files)
    e.target.value = ''
  }

  const handleFileDrop = async (e: React.DragEvent) => {
    e.stopPropagation(); e.preventDefault()
    try { await ingestFiles(e.dataTransfer?.files) }
    catch (err) { toast.error(`Could not add dropped files: ${err instanceof Error ? err.message : err}`); notifyError(err, 'file-drop') }
  }

  const removeFile = (index: number) => {
    clearValidation()
    setSelectedFiles((prev) => prev.filter((_, i) => i !== index))
  }

  const clearFiles = () => {
    clearValidation()
    setSelectedFiles([])
    setBulkResults(null)
    setJobId(''); setJobStatus(''); setLastUpdated(null); setResultError('')
  }

  const validateAll = (opts: BulkOptions): string[] => {
    const errors: string[] = []
    if (!opts.renameFiles && !opts.removeMetadata && !opts.optimizeFiles && !opts.compressFiles)
      errors.push('Enable at least one processing option (rename, metadata removal, optimization, or compression).')
    if (opts.renameFiles) {
      const pErr = validatePattern(opts.pattern || '', { preserveOriginalName: opts.renameOptions?.preserveOriginalName })
      if (pErr) errors.push(pErr)
      if (opts.renameOptions?.sequentialNaming?.enabled) {
        const base = opts.renameOptions.sequentialNaming.baseName?.trim()
        if (!base && !opts.renameOptions.preserveOriginalName)
          errors.push('When sequential naming is enabled, provide a base name or enable preserve original name.')
      }
    }
    if (!opts.allowedTypes?.length) errors.push('Allowed content types list cannot be empty.')
    return errors
  }

  const buildRequest = (opts: BulkOptions) => {
    return new advanced_file_operations.BulkProcessingRequest({
      files: selectedFiles.map((f) => new advanced_file_operations.BulkProcessingFile({
        filename: f.name, contentBase64: f.base64, contentType: f.type || '', size: f.size,
      })),
      options: new advanced_file_operations.BulkProcessingOptions({
        renameFiles: opts.renameFiles, removeMetadata: opts.removeMetadata,
        optimizeFiles: opts.optimizeFiles, compressFiles: opts.compressFiles,
        pattern: opts.pattern, namer: opts.namer, renameOptions: opts.renameOptions,
        allowedTypes: opts.allowedTypes?.length ? opts.allowedTypes : [...DEFAULT_ALLOWED_TYPES],
        maxFileSize: opts.maxFileSize,
      }),
    })
  }

  const processSelectedFiles = async () => {
    if (!selectedFiles.length) { toast('Add at least one file to process.'); return }
    const oversize = selectedFiles.filter((f) => f.size > bulkOptions.maxFileSize)
    if (oversize.length) {
      const message = `${oversize.length} file(s) exceed the max size of ${formatBytes(bulkOptions.maxFileSize)}.`
      toast.error(message); notifyError(message, 'file-validation'); return
    }
    const normalizedPattern = normalizePatternInput(bulkOptions.pattern)
    const trimmedBase = bulkOptions.renameOptions.sequentialNaming.baseName?.trim() ?? ''
    const candidateOptions: BulkOptions = {
      ...bulkOptions, pattern: normalizedPattern,
      renameOptions: { ...bulkOptions.renameOptions, sequentialNaming: { ...bulkOptions.renameOptions.sequentialNaming, baseName: trimmedBase } },
    }
    const configErrors = validateAll(candidateOptions)
    setValidationErrors(configErrors)
    if (configErrors.length) { setPatternTouched(true); toast.error(configErrors[0]); return }
    setBulkOptions(candidateOptions)
    setValidationErrors([])
    setIsProcessing(true); setResultError(''); setBulkResults(null)
    try {
      const response = await ProcessBulkFiles(buildRequest(candidateOptions)) as BulkProcessingResponse
      setBulkResults(response)
      setJobId(response?.jobId || '')
      setJobStatus(response ? ((response.failureCount ?? 0) > 0 ? 'completed_with_errors' : 'completed') : '')
      setLastUpdated(new Date())
      toast.success('Advanced processing complete.')
    } catch (err) {
      const msg = err instanceof Error ? err.message : String(err)
      setResultError(msg); toast.error(`Processing failed: ${msg}`); notifyError(err, 'bulk-processing')
    } finally { setIsProcessing(false) }
  }

  const refreshJobStatus = async () => {
    if (!jobId) return
    setIsRefreshing(true); setResultError('')
    try {
      const job = await GetBulkProcessingJob(jobId) as Record<string, unknown>
      setJobStatus((job?.status as string) || jobStatus)
      setLastUpdated(new Date())
      if (job) {
        const existingByFilename = new Map((bulkResults?.results || []).map((r) => [r.filename, r]))
        setBulkResults({
          jobId: job.id as string,
          totalFiles: ((job.files as unknown[]) ?? []).length,
          successCount: ((job.results as ProcessingResult[]) ?? []).filter((i) => i.success).length,
          failureCount: ((job.results as ProcessingResult[]) ?? []).filter((i) => !i.success).length,
          durationMs: (job.durationMs as number) ?? bulkResults?.durationMs ?? 0,
          results: ((job.results as ProcessingResult[]) ?? []).map((item) => ({
            filename: item.filename, newName: item.newName, success: item.success,
            error: item.error, action: item.action, contentType: item.contentType,
            contentBase64: existingByFilename.get(item.filename)?.contentBase64,
          })),
        })
      }
    } catch (err) {
      const msg = err instanceof Error ? err.message : String(err)
      setResultError(msg); toast.error(`Could not refresh job: ${msg}`); notifyError(err, 'job-refresh')
    } finally { setIsRefreshing(false) }
  }

  const downloadProcessedFile = (result: ProcessingResult) => {
    if (!result?.contentBase64) { toast('Run a new batch to retrieve downloads for this job.'); return }
    const link = document.createElement('a')
    link.href = `data:${result.contentType || 'application/octet-stream'};base64,${result.contentBase64}`
    link.download = result.newName || result.filename || 'processed-file'
    document.body.appendChild(link); link.click(); document.body.removeChild(link)
  }

  const handleMaxFileSizeBlur = () => {
    const value = Number(maxFileSizeInput)
    const normalised = Number.isFinite(value) && value > 0 ? value : 50
    setMaxFileSizeInput(normalised)
    setBulkOptions((prev) => ({ ...prev, maxFileSize: normalised * 1024 * 1024 }))
  }

  return (
    <div className="config-card">
      <div className="card-header">
        <h3>🧰 Bulk Processing Toolkit</h3>
        <p>Run ad-hoc files through the advanced operations pipeline.</p>
      </div>
      <div className="card-content bulk-operations">
        {/* File Upload */}
        <div className="file-picker">
          <label
            className="upload-area"
            onDragOver={(e) => e.preventDefault()}
            onDragEnter={(e) => e.preventDefault()}
            onDrop={handleFileDrop}
          >
            <input type="file" multiple onChange={handleFileSelection} />
            <span className="upload-icon">📤</span>
            <span className="upload-text">
              Drop files here or <strong>click to browse</strong>
              <small>Up to {MAX_FILES_PER_BATCH} files per batch</small>
            </span>
          </label>
          {selectedFiles.length > 0 && (
            <>
              <div className="file-summary">
                <span>{selectedFiles.length} files selected • {formatBytes(totalSelectedSize)}</span>
                <button className="link-btn" onClick={clearFiles}>Clear selection</button>
              </div>
              <ul className="file-list">
                {selectedFiles.map((file, index) => (
                  <li key={index} className={!isAllowedType(file.type) ? 'disallowed' : ''}>
                    <div className="file-meta">
                      <strong>{file.name}</strong>
                      <small>{file.type || 'Unknown'} • {formatBytes(file.size)}</small>
                      {!isAllowedType(file.type) && <span className="badge warn">Type not in allowed list</span>}
                    </div>
                    <button className="link-btn" onClick={() => removeFile(index)}>Remove</button>
                  </li>
                ))}
              </ul>
            </>
          )}
        </div>

        {/* Options */}
        <div className="options-grid">
          {([
            { key: 'renameFiles', label: 'Rename Files', desc: 'Apply naming templates or sequential patterns' },
            { key: 'removeMetadata', label: 'Remove Metadata', desc: 'Strip EXIF and embedded metadata (requires ExifTool)' },
            { key: 'optimizeFiles', label: 'Optimize Files', desc: 'Run format-specific optimisation for supported types' },
            { key: 'compressFiles', label: 'Compress Output', desc: 'Apply lossless compression to processed files' },
          ] as const).map(({ key, label, desc }) => (
            <div key={key} className="option-toggle">
              <Toggle
                checked={bulkOptions[key]}
                onChange={(checked) => { clearValidation(); setBulkOptions((prev) => ({ ...prev, [key]: checked })) }}
                ariaLabel={label}
              />
              <span className="toggle-label">
                <strong>{label}</strong>
                <small>{desc}</small>
              </span>
            </div>
          ))}
        </div>

        {/* Rename Settings */}
        {bulkOptions.renameFiles && (
          <div className="rename-settings">
            <h4>Rename Settings</h4>
            <div className="rename-grid">
              <label>
                <span>Pattern</span>
                <input
                  type="text"
                  value={bulkOptions.pattern}
                  className={patternTouched && !!patternError ? 'input-error' : ''}
                  onChange={(e) => { clearValidation(); setBulkOptions((prev) => ({ ...prev, pattern: e.target.value })) }}
                  onBlur={() => setPatternTouched(true)}
                  placeholder="Example: random | timestamp | seq:IMG:1:3"
                />
                <small className="muted">Leave blank to keep names, or use presets like random, timestamp, seq:Base:Start:Pad, regex:find:replace.</small>
                {patternTouched && patternError && <small className="field-error">{patternError}</small>}
              </label>
              <label>
                <span>Namer ID</span>
                <input
                  type="text"
                  value={bulkOptions.namer}
                  onChange={(e) => setBulkOptions((prev) => ({ ...prev, namer: e.target.value }))}
                  placeholder="random | template | sequential"
                />
              </label>
              <label className="inline-option">
                <input
                  type="checkbox"
                  checked={bulkOptions.renameOptions.preserveOriginalName}
                  onChange={(e) => updateRenameOptions({ preserveOriginalName: e.target.checked })}
                />
                <span>Preserve original base name</span>
              </label>
              <label className="inline-option">
                <input
                  type="checkbox"
                  checked={bulkOptions.renameOptions.addTimestamp}
                  onChange={(e) => updateRenameOptions({ addTimestamp: e.target.checked })}
                />
                <span>Append timestamp</span>
              </label>
              <label className="inline-option">
                <input
                  type="checkbox"
                  checked={bulkOptions.renameOptions.addRandomId}
                  onChange={(e) => updateRenameOptions({ addRandomId: e.target.checked })}
                />
                <span>Append random suffix</span>
              </label>
            </div>
            <div className="sequential-card">
              <div className="sequential-header">
                <label className="inline-option">
                  <input
                    type="checkbox"
                    checked={bulkOptions.renameOptions.sequentialNaming.enabled}
                    onChange={(e) => updateSequentialOptions({ enabled: e.target.checked })}
                  />
                  <span>Enable sequential naming</span>
                </label>
              </div>
              {bulkOptions.renameOptions.sequentialNaming.enabled && (
                <div className="sequential-grid">
                  <label>
                    <span>Base name</span>
                    <input
                      type="text"
                      value={bulkOptions.renameOptions.sequentialNaming.baseName}
                      onChange={(e) => { clearValidation(); updateSequentialOptions({ baseName: e.target.value }) }}
                    />
                  </label>
                  <label>
                    <span>Start index</span>
                    <input
                      type="number"
                      min="0"
                      value={bulkOptions.renameOptions.sequentialNaming.startIndex}
                      onChange={(e) => { clearValidation(); updateSequentialOptions({ startIndex: Number(e.target.value) }) }}
                    />
                  </label>
                  <label>
                    <span>Pad length</span>
                    <input
                      type="number"
                      min="1"
                      value={bulkOptions.renameOptions.sequentialNaming.padLength}
                      onChange={(e) => { clearValidation(); updateSequentialOptions({ padLength: Number(e.target.value) }) }}
                    />
                  </label>
                  <label className="inline-option">
                    <input
                      type="checkbox"
                      checked={bulkOptions.renameOptions.sequentialNaming.keepExtension}
                      onChange={(e) => updateSequentialOptions({ keepExtension: e.target.checked })}
                    />
                    <span>Keep original extension</span>
                  </label>
                </div>
              )}
            </div>
          </div>
        )}

        {/* Constraints */}
        <div className="constraints-grid">
          <div className="info-field">
            <span className="field-label">Allowed content types</span>
            <div className="allowed-types" aria-live="polite">{allowedTypesDisplay}</div>
            <small className="muted">Configured MIME types (comma-separated)</small>
          </div>
          <label>
            <span>Max file size (MB)</span>
            <input
              type="number"
              min="1"
              max="200"
              value={maxFileSizeInput}
              onChange={(e) => setMaxFileSizeInput(Number(e.target.value))}
              onBlur={handleMaxFileSizeBlur}
            />
            <small>Files larger than this limit are skipped</small>
          </label>
        </div>

        {/* Validation Errors */}
        {validationErrors.length > 0 && (
          <div className="validation-errors" role="alert">
            {validationErrors.map((err, i) => <div key={i}>{err}</div>)}
          </div>
        )}

        {/* Action Buttons */}
        <div className="actions">
          <Button
            variant="primary"
            onClick={processSelectedFiles}
            disabled={isProcessing || !selectedFiles.length}
          >
            {isProcessing ? <><span className="spinner" />Processing…</> : 'Run Advanced Operations'}
          </Button>
          <Button
            variant="ghost"
            onClick={refreshJobStatus}
            disabled={!jobId || isRefreshing}
          >
            {isRefreshing ? <><span className="spinner" />Refreshing…</> : 'Refresh last job'}
          </Button>
        </div>

        {/* Result Error */}
        {resultError && <div className="result-error">⚠️ {resultError}</div>}

        {/* Results */}
        {bulkResults && (
          <div className="results-panel">
            <div className="results-header">
              <div><strong>Job ID:</strong> {bulkResults.jobId}</div>
              <div className="results-meta">
                <span>Status: {jobStatus || 'completed'}</span>
                <span>Success: {bulkResults.successCount} • Failures: {bulkResults.failureCount}</span>
                <span>Duration: {formatDuration(bulkResults.durationMs)}</span>
                {lastUpdated && <span>Updated: {lastUpdated.toLocaleTimeString()}</span>}
              </div>
            </div>
            <table className="results-table">
              <thead>
                <tr><th>File</th><th>Action</th><th>Status</th><th className="min">Download</th></tr>
              </thead>
              <tbody>
                {(bulkResults.results || []).map((result, i) => (
                  <tr key={i} className={!result.success ? 'has-error' : ''}>
                    <td>
                      <div className="result-name">
                        <strong>{result.newName || result.filename}</strong>
                        <small>{result.filename}</small>
                      </div>
                    </td>
                    <td>{result.action || '—'}</td>
                    <td>
                      {result.success
                        ? <span className="badge success">Success</span>
                        : <span className="badge warn">{result.error || 'Failed'}</span>}
                    </td>
                    <td className="min">
                      {result.success
                        ? <button className="link-btn" onClick={() => downloadProcessedFile(result)}>Download</button>
                        : '—'}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  )
}
