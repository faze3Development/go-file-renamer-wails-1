<script lang="ts">
  import { onMount } from 'svelte';
  import { settings, themes, applyTheme, LOG_RETENTION_DEFAULT, BACKEND_VERBOSITY_DEFAULT } from './stores';
  import toast from 'svelte-french-toast';
  import { Toaster } from 'svelte-french-toast';
  import Sidebar from './components/Sidebar.svelte';
  import SettingsModal from './components/SettingsModal.svelte';
  import AdvancedView from './components/views/AdvancedView.svelte';
  import AdvancedOperationsView from './components/views/AdvancedOperationsView.svelte';
  import ConfigurationView from './components/ConfigurationView.svelte';
  import MonitoringView from './components/MonitoringView.svelte';
  import ErrorBoundary from './components/ErrorBoundary.svelte';
  import LoadingBoundary from './components/LoadingBoundary.svelte';
  import { EventsOn, EventsOff, OnFileDrop, OnFileDropOff } from '../wailsjs/runtime/runtime.js';
  import { LoadProfiles, GetNamerInfo, GetActionInfo, GetPatternInfo, SaveProfile, DeleteProfile, SelectActionDirectory, SelectDirectory, StartWatching, StopWatching } from '../wailsjs/go/main/App.js';
  import { CircleAlert, RefreshCw, X, FileCog, Activity, Cpu, Folder, FolderOpen, Eye, FileText, Cog, Play, Square, BookText } from 'lucide-svelte';
  import { logStore, pushLogEntry, fromErrorPayload } from './lib/logStore';
  import { isLogEntryPayload, isStatsPayload, type LogEntryPayload, type StatsPayload } from './types/events';

  // --- Utility Functions ---
  function checkRuntimeAvailable() {
    if (typeof window === "undefined" || !window.runtime) {
      toast("Runtime not available - please wait for app to fully load", {
        duration: 4000,
        position: "top-right",
      });
      return false;
    }
    return true;
  }

  function handleError(error, context = "") {
    hasError = true;
    errorMessage = (context ? `[${context}] ` : "") + (error?.message || error);
    console.error(errorMessage);
    pushLogEntry({
      message: errorMessage,
      severity: "error",
      source: WATCHER_SOURCE,
      details: error,
      context,
    });
  }

  function clearError() {
    hasError = false;
    errorMessage = "";
  }


  const WATCHER_SOURCE = "watcher";
  const ADVANCED_OPERATIONS_SOURCE = "advanced-operations";
  const severityToLevel = {
    debug: "DEBUG",
    info: "INFO",
    warn: "WARN",
    error: "ERROR",
    fatal: "FATAL",
  };
  const severityRank = {
    debug: 0,
    info: 1,
    warn: 2,
    error: 3,
    fatal: 4,
  };

  let backendVerbosityPrefs = { ...BACKEND_VERBOSITY_DEFAULT };

  // No legacy watcher-level mapping required; logs are canonicalized upstream.

  function mapBackendSourceToPreferenceKey(source) {
    switch (source) {
      case WATCHER_SOURCE:
        return "watcher";
      case ADVANCED_OPERATIONS_SOURCE:
        return "advancedOperations";
      default:
        return "global";
    }
  }

  function shouldIngestBackendLog(entry) {
    const severity = entry.severity ?? "info";
    const rank = severityRank[severity] ?? severityRank.info;

    if (rank >= severityRank.warn) {
      return true;
    }

    const prefKey = mapBackendSourceToPreferenceKey(entry.source);
    const sourceEnabled = backendVerbosityPrefs?.[prefKey];

    if (!sourceEnabled) {
      return false;
    }

    if (severity === "debug") {
      return Boolean(backendVerbosityPrefs?.global);
    }

    return true;
  }

  function scheduleLogScroll() {
    setTimeout(() => {
      try {
        if (logContainer) {
          logContainer.scrollTop = logContainer.scrollHeight;
        }
      } catch (scrollError) {
        console.warn("Log scroll error:", scrollError);
      }
    }, 50);
  }

  function ingestBackendLog(payload: unknown): void {
    // Type guard: validate payload is an object with expected structure
    if (!isLogEntryPayload(payload)) {
      console.warn('Invalid log entry payload received:', payload);
      return;
    }

    const entry: LogEntryPayload = {
      message: payload.message,
      severity: payload.severity ?? "info",
      source: payload.source ?? "backend",
      context: payload.context,
      details: payload.details,
      metadata: payload.metadata,
      timestamp: payload.timestamp ?? new Date().toISOString(),
      id: payload.id,
    };

    if (!shouldIngestBackendLog(entry)) {
      return;
    }

    pushLogEntry(entry);
    scheduleLogScroll();
  }

  // Legacy watcher ingestion removed. The backend/logStore should emit canonical LogEntryPayload entries.

  $: monitoringLogs = $logStore
    .filter((entry) => (entry.source ?? WATCHER_SOURCE) === WATCHER_SOURCE)
    .map((entry) => ({
      // Normalize to LogEntryPayload shape used across the app
      message: entry.message,
      severity: entry.severity,
      source: entry.source ?? WATCHER_SOURCE,
      context: entry.context,
      details: entry.details,
      metadata: entry.metadata,
      timestamp: entry.timestamp,
      id: entry.id,
    }));




  function retryInitialization() {
    isInitialized = false;
    clearError();
    loadInitialData();
  }

  async function handleStart() {
    if (!watchPathDisplay) {
      toast("Please select a directory to watch first.", {
        duration: 3000,
        position: "top-right",
      });
      return;
    }

    if (!checkRuntimeAvailable()) return;

    isBusy = true; // Set busy state
  logStore.clear(); // Clear log history on start
    stats = {
      // Reset stats on start for a clean slate
      scanned: 0,
      renamed: 0,
      skipped: 0,
      errors: 0,
    };

    try {
      await StartWatching(config);
      isWatching = true;
      toast.success("File watching started successfully", {
        duration: 3000,
        position: "top-right",
      });
    } catch (error) {
      console.error("Failed to start watcher:", error);
      const msg = (error && error.message) ? error.message : String(error);
      // Detect the common backend message when a watcher is already running
      if (msg.toLowerCase().includes('already running')) {
        // Show modal allowing user to attach or force-stop the running watcher
        conflictError = error;
        showWatcherConflictModal = true;
        pushLogEntry({
          message: `Watcher already running: ${msg}`,
          severity: "warn",
          source: WATCHER_SOURCE,
          details: error,
        });
        toast(`Watcher already running`, { duration: 5000, position: 'top-right' });
      } else {
        toast.error(`Failed to start watching: ${msg}`, {
          duration: 5000,
          position: "top-right",
        });
        pushLogEntry({
          message: `Failed to start watcher: ${msg}`,
          severity: "error",
          source: WATCHER_SOURCE,
          details: error,
        });
      }
      isBusy = false; // Reset busy state on failure
    }
  }

  async function handleStop() {
    if (!checkRuntimeAvailable()) return;

    try {
      isBusy = true; // Set busy state
      await StopWatching();
      isWatching = false;
      pushLogEntry({
        message: "Watcher stop requested",
        severity: "info",
        source: WATCHER_SOURCE,
      });
    } catch (error) {
      console.error("Failed to stop watcher:", error);
      toast.error(`Failed to stop watching: ${error.message || error}`, {
        duration: 5000,
        position: "top-right",
      });
      isBusy = false; // Reset busy state on failure
      pushLogEntry({
        message: `Failed to stop watcher: ${error.message || error}`,
        severity: "error",
        source: WATCHER_SOURCE,
        details: error,
      });
    }
  }

  let config = {
    WatchPaths: [],
    Recursive: true,
    DryRun: false,
    NamePattern:
      "(?i)^(?:untitled|screenshot(?:[\s_-]\d{4}-\d{2}-\d{2})?|untitiled)(?:[\s_-]*(?:\(\d+\)|\d+))?$",
    TemplateString: "{original}-{date}",
    DateTimeFormat: "2006-01-02_15-04-05",
    RandomLength: 12,
    Settle: 750, // in ms
    SettleTimeout: 10, // in seconds
    Retries: 6,
    NoInitialScan: false,
    NamerID: "random",
    ActionID: "none",
    ActionConfig: { destinationPath: "" },
  };

  let watchPathDisplay = "";
  let isWatching = false;
  let isBusy = false; // New state to manage start/stop transitions
  let showWatcherConflictModal = false;
  let conflictError: any = null;
  let modalBusy = false;
  /** @type {HTMLElement} */
  let logContainer;
  let stats = {
    scanned: 0,
    renamed: 0,
    skipped: 0,
    errors: 0,
  };

  let profiles = {};
  let availableNamers = [];
  let availableActions = [];
  let availablePatterns = [];
  let selectedProfile = "";
  let selectedPatternID = "default_untitled";

  interface AppSettings {
    showActivityLog?: boolean;
    showStats?: boolean;
    compactMode?: boolean;
    [key: string]: any;
  }

  // Settings state
  let showSettingsModal = false;
  let currentSettings: AppSettings = {};
  let currentView = "configuration"; // 'configuration', 'monitoring', 'advanced', 'advancedOperations'
  // Error handling state
  let hasError = false;
  let errorMessage = "";
  let isLoading = true;
  let isInitialized = false;

  // Guard to ensure we only teardown listeners once (HMR / multiple unmount safety)
  let eventsListenersRemoved = false;
  // Deduplication timestamp for watcherStopped to avoid repeated toasts
  let lastWatcherStoppedAt = 0;



  // --- Lifecycle & Event Handlers ---
  onMount(() => {
    // Subscribe to settings changes with error handling
    const unsubscribe = settings.subscribe((value) => {
      try {
        currentSettings = value;
        if (value?.theme) {
          applyTheme(value.theme);
        }

        const nextLimit = value?.logRetentionLimit ?? LOG_RETENTION_DEFAULT;
        logStore.setLimit(nextLimit);

        backendVerbosityPrefs = {
          ...BACKEND_VERBOSITY_DEFAULT,
          ...(value?.backendVerbosity ?? {}),
        };
      } catch (error) {
        handleError(error, "Settings subscription");
      }
    });

    // Wait for Wails runtime to be available with comprehensive error handling
    const initializeWailsListeners = () => {
      try {
        if (
          typeof window !== "undefined" &&
          window.runtime &&
          window.runtime.EventsOnMultiple
        ) {
          try {
            // Structured backend log stream (preferred)
            EventsOn("log:entry", (payload: unknown) => {
              try {
                ingestBackendLog(payload);
              } catch (logError) {
                console.warn("Structured log ingest error:", logError);
              }
            });

            // Legacy watcher logs for compatibility during migration
            EventsOn("logEntry", (log: unknown) => {
              try {
                ingestLegacyWatcherLog(log);
              } catch (legacyError) {
                console.warn("Legacy log ingest error:", legacyError);
              }
            });

            // Listen for when the watcher starts from the backend
            EventsOn("watcherStarted", () => {
              try {
                isWatching = true;
                isBusy = false; // No longer busy, watcher is running
              } catch (startError) {
                console.warn("Watcher start event error:", startError);
              }
            });

            // Listen for when the watcher stops from the backend
            EventsOn("watcherStopped", () => { // Corrected event name
              try {
                const now = Date.now();
                // Deduplicate rapid successive watcherStopped events (2s window)
                if (now - lastWatcherStoppedAt < 2000) {
                  console.debug('Ignored duplicate watcherStopped event (within 2s)');
                  return;
                }
                lastWatcherStoppedAt = now;

                isWatching = false; // Update state based on backend confirmation
                isBusy = false; // No longer busy, watcher is stopped
                toast("File watching stopped", {
                  duration: 3000,
                  position: "top-right",
                });
              } catch (stopError) {
                console.warn("Watcher stop error:", stopError);
              }
            });

            // Listen for stats updates from the Go backend
            EventsOn("statsUpdated", (newStats: unknown) => { // Corrected event name
              try {
                if (isStatsPayload(newStats)) {
                  stats = { ...stats, ...newStats };
                }
              } catch (statsError) {
                console.warn("Stats update error:", statsError);
              }
            });

            // Setup Drag and Drop with error handling
            OnFileDrop((x, y, paths) => {
              try {
                if (isWatching) return; // Don't allow drops while watching
                if (paths && paths.length > 0) {
                  // Use the first dropped path
                  watchPathDisplay = paths[0];
                  config.WatchPaths = [paths[0]];
                }
              } catch (dropError) {
                handleError(dropError, "File drop");
              }
            }, false); // `false` means the whole window is a drop target

            console.log("Wails event listeners initialized successfully");
          } catch (error) {
            handleError(error, "Wails event listener setup");
          }
        } else {
          // Retry after a short delay if runtime isn't ready
          setTimeout(initializeWailsListeners, 100);
        }
      } catch (error) {
        handleError(error, "Wails initialization");
      }
    };

    // Initialize Wails listeners
    initializeWailsListeners();

    // Load initial data from backend
    loadInitialData();

    // Cleanup function
    return () => {
      try {
        if (unsubscribe) unsubscribe();
        // Remove Wails event listeners to avoid duplicate handlers on HMR/frontend reloads
        if (!eventsListenersRemoved && typeof window !== 'undefined' && window.runtime && window.runtime.EventsOff) {
          try {
            EventsOff('log:entry', 'logEntry', 'watcherStarted', 'watcherStopped', 'statsUpdated');
            eventsListenersRemoved = true;
          } catch (e) {
            console.warn('Failed to remove Wails event listeners during cleanup:', e);
          }
        }
        // Remove drag/drop handlers as well
        try {
          if (typeof window !== 'undefined' && window.runtime && window.runtime.OnFileDropOff) {
            OnFileDropOff();
          }
        } catch (e) {
          console.warn('Failed to remove OnFileDrop handlers during cleanup:', e);
        }
      } catch (err) {
        console.warn('Error during App cleanup:', err);
      }
    };
  });

  /**
   * Wraps a promise with a timeout.
   * @param {Promise<T>} promise The promise to wrap.
   * @param {number} ms The timeout in milliseconds.
   * @returns {Promise<T>}
   */
  function withTimeout(promise, ms = 5000) {
    return new Promise((resolve, reject) => {
      const timer = setTimeout(() => reject(new Error('Request timed out')), ms);
      promise.then(resolve, reject).finally(() => clearTimeout(timer));
    });
  }
  async function loadInitialData() {
    try {
      isLoading = true;
      clearError();

      // Check if Wails runtime is available before making API calls
      if (typeof window !== "undefined" && window.runtime && window.go) {
        try {
          const [loadedProfiles, loadedNamers, loadedActions, loadedPatterns] =
            await Promise.allSettled([
              withTimeout(LoadProfiles()),
              withTimeout(GetNamerInfo()),
              withTimeout(GetActionInfo()),
              withTimeout(GetPatternInfo()),
            ]);

          // Handle each result individually to prevent one failure from affecting others
          profiles =
            loadedProfiles.status === "fulfilled"
              ? loadedProfiles.value || {}
              : {};
          availableNamers =
            loadedNamers.status === "fulfilled" ? loadedNamers.value || [] : [];
          availableActions =
            loadedActions.status === "fulfilled"
              ? loadedActions.value || []
              : [];
          availablePatterns =
            loadedPatterns.status === "fulfilled"
              ? loadedPatterns.value || []
              : [];

          // Log any failures
          if (loadedProfiles.status === "rejected")
            console.warn("Failed to load profiles:", loadedProfiles.reason);
          if (loadedNamers.status === "rejected")
            console.warn("Failed to load namers:", loadedNamers.reason);
          if (loadedActions.status === "rejected")
            console.warn("Failed to load actions:", loadedActions.reason);
          if (loadedPatterns.status === "rejected")
            console.warn("Failed to load patterns:", loadedPatterns.reason);

          // Set initial pattern selection if available
          if (availablePatterns.length > 0) {
            try {
              const matchingPattern = availablePatterns.find(
                (p) => p.regex === config.NamePattern,
              );
              if (matchingPattern) {
                selectedPatternID = matchingPattern.id;
              } else {
                selectedPatternID = availablePatterns[0].id; // Default to first pattern
                config.NamePattern = availablePatterns[0].regex;
              }
            } catch (patternError) {
              console.warn("Pattern selection error:", patternError);
            }
          }

          console.log("Initial data loaded successfully");
          isInitialized = true;
          return true;
        } catch (e) {
          throw new Error(`API call failed: ${e.message || e}`);
        }
      } else {
        // Fallback data for when backend is not available (browser/development mode)
        console.log(
          "Wails runtime not available, using fallback data for browser development",
        );

        // Provide consistent fallback data that matches expected backend structure
        availablePatterns = [
          {
            id: "default_untitled",
            name: "Default Untitled/Screenshot",
            regex:
              "(?i)^(?:untitled|screenshot(?:[\\s_-]\\d{4}-\\d{2}-\d{2})?|untitled)(?:[\\s_-]*(?:\\(\\d+\\)|\\d+))?$",
            description: "Matches untitled files and screenshots",
          },
        ];

        availableNamers = [
          {
            id: "random",
            name: "Random Name",
            description: "Generate random alphanumeric names",
          },
          {
            id: "template",
            name: "Template Based",
            description:
              "Use template with placeholders like {original}-{date}",
          },
          {
            id: "timestamp",
            name: "Timestamp",
            description: "Use current timestamp as filename",
          },
        ];

        availableActions = [
          {
            id: "none",
            name: "No Action",
            description: "Just rename the file in place",
          },
          {
            id: "move",
            name: "Move File",
            description: "Move file to another directory after renaming",
          },
          {
            id: "copy",
            name: "Copy File",
            description: "Copy file to another directory after renaming",
          },
          {
            id: "advanced_operations",
            name: "Advanced File Operations",
            description:
              "Run the advanced processing pipeline with metadata tooling",
          },
        ];

        profiles = {};
        selectedPatternID = "default_untitled";

        console.log(
          "Fallback data loaded successfully for browser environment",
        );
        isInitialized = true;
        return true;
      }
    } catch (error) {
      handleError(error, "Loading initial data");
      return false;
    } finally {
      isLoading = false;
    }
  }

  // --- UI Handlers ---

  // When the pattern preset changes, update the actual regex in the config
  function handlePatternSelect(event) {
    try {
      const id = event?.target?.value;
      if (!id) return;

      selectedPatternID = id;
      if (id !== "custom" && availablePatterns.length > 0) {
        const selectedPattern = availablePatterns.find((p) => p.id === id);
        if (selectedPattern && selectedPattern.regex) {
          config.NamePattern = selectedPattern.regex;
        }
      }
    } catch (error) {
      handleError(error, "Pattern selection");
    }
  }

  function handleProfileSelect(event) {
    try {
      const profileName = event?.target?.value;
      selectedProfile = profileName;

      if (profileName && profiles && profiles[profileName]) {
        // Create a new object to avoid reactivity issues and merge with defaults
        const profileData = profiles[profileName];
        if (profileData && typeof profileData === "object") {
          config = {
            ...config,
            ...profileData,
            ActionConfig: {
              ...config.ActionConfig,
              ...(profileData.ActionConfig || {}),
            },
          };
          watchPathDisplay =
            config.WatchPaths && config.WatchPaths.length > 0
              ? config.WatchPaths[0]
              : "";

          // Also update the selected pattern dropdown
          if (availablePatterns.length > 0) {
            const matchingPattern = availablePatterns.find(
              (p) => p.regex === config.NamePattern,
            );
            if (matchingPattern) {
              selectedPatternID = matchingPattern.id;
            } else {
              selectedPatternID = "custom";
            }
          }

          toast.success(`Profile '${profileName}' loaded successfully`, {
            duration: 3000,
            position: "top-right",
          });
        }
      }
    } catch (error) {
      handleError(error, "Profile selection");
    }
  }

  async function handleSaveProfile() {
    if (!checkRuntimeAvailable()) return;

    const name = prompt(
      "Enter a name for this profile:",
      selectedProfile || "New Profile",
    );
    if (name) {
      try {
        await SaveProfile(name, config);
        await loadInitialData(); // Reload profiles from backend
        selectedProfile = name; // Set the dropdown to the new profile
        toast.success(`Profile '${name}' saved successfully`, {
          duration: 3000,
          position: "top-right",
        });
      } catch (e) {
        console.error("Failed to save profile:", e);
        toast.error(`Failed to save profile: ${e.message || e}`, {
          duration: 5000,
          position: "top-right",
        });
      }
    }
  }

  async function handleDeleteProfile() {
    if (!selectedProfile) {
      toast("No profile selected to delete", {
        duration: 3000,
        position: "top-right",
      });
      return;
    }

    if (!checkRuntimeAvailable()) return;

    if (
      confirm(
        `Are you sure you want to delete the profile '${selectedProfile}'?`,
      )
    ) {
      try {
        await DeleteProfile(selectedProfile);
        selectedProfile = "";
        await loadInitialData(); // Reload profiles
        toast.success("Profile deleted successfully", {
          duration: 3000,
          position: "top-right",
        });
      } catch (e) {
        console.error("Failed to delete profile:", e);
        toast.error(`Failed to delete profile: ${e.message || e}`, {
          duration: 5000,
          position: "top-right",
        });
      }
    }
  }

  // --- Settings Functions ---

  function openSettings() {
    showSettingsModal = true;
  }

  function closeSettings() {
    showSettingsModal = false;
  }

  function handleUpdateConfig(event) {
    config[event.detail.key] = event.detail.value;
  }



  function updateTheme(themeName) {
    try {
      if (!themeName || !themes[themeName]) {
        throw new Error(`Invalid theme: ${themeName}`);
      }
      settings.updateSetting("theme", themeName);
      toast.success("Theme updated successfully", {
        duration: 3000,
        position: "top-right",
      });
    } catch (error) {
      handleError(error, "Theme update");
    }
  }

  function toggleActivityLog() {
    try {
      settings.updateSetting(
        "showActivityLog",
        !currentSettings.showActivityLog,
      );
    } catch (error) {
      handleError(error, "Toggle activity log");
    }
  }

  function toggleStats() {
    try {
      settings.updateSetting("showStats", !currentSettings.showStats);
    } catch (error) {
      handleError(error, "Toggle statistics");
    }
  }

  function toggleCompactMode() {
    try {
      settings.updateSetting("compactMode", !currentSettings.compactMode);
    } catch (error) {
      handleError(error, "Toggle compact mode");
    }
  }

  // --- Core App Functions ---

  async function handleSelectActionDirectory() {
    if (!checkRuntimeAvailable()) return;

    try {
      const selectedPath = await SelectActionDirectory();
      if (selectedPath) {
        // Ensure ActionConfig exists and update the destinationPath
        if (!config.ActionConfig) {
          config.ActionConfig = { destinationPath: "" };
        }
        config.ActionConfig.destinationPath = selectedPath;
        // Force reactivity update
        config = config;
      }
    } catch (error) {
      console.error("Error selecting action directory:", error);
      toast.error(`Failed to select directory: ${error.message || error}`, {
        duration: 5000,
        position: "top-right",
      });
    }
  }

  async function handleSelectDirectory() {
    if (!checkRuntimeAvailable()) return;

    try {
      const selectedPath = await SelectDirectory();
      if (selectedPath) {
        watchPathDisplay = selectedPath;
        config.WatchPaths = [selectedPath];
      }
    } catch (error) {
      console.error("Error selecting directory:", error);
      toast.error(`Failed to select directory: ${error.message || error}`, {
        duration: 5000,
        position: "top-right",
      });
    }
  }

  function handleAdvancedOperationsError(event) {
    const payload = event?.detail || {};
    const context = payload.context || "advanced operations";
    const defaultMessage = payload.message || "Unexpected advanced operations error";

    if (payload.severity === "fatal") {
      const fatalDetail = payload.details instanceof Error ? payload.details : defaultMessage;
      handleError(fatalDetail, context);
      return;
    }

    const logged = payload.details instanceof Error ? payload.details : defaultMessage;
    console.error(`[${context}]`, logged);
    pushLogEntry({
      ...fromErrorPayload(payload),
      source: ADVANCED_OPERATIONS_SOURCE,
    });
  }

  // --- Watcher conflict actions (UI-only handlers) ---
  async function attachToRunningWatcher() {
    // Optimistic attach: mark UI as watching and close modal.
    // This assumes the backend watcher is running and will continue to emit events.
    showWatcherConflictModal = false;
    conflictError = null;
    isWatching = true;
    pushLogEntry({
      message: 'Attached to running watcher (UI only)',
      severity: 'info',
      source: WATCHER_SOURCE,
    });
    toast.success('Attached to running watcher', { duration: 3000, position: 'top-right' });
  }

  async function forceStopAndStart() {
    modalBusy = true;
    try {
      await StopWatching();
      pushLogEntry({ message: 'Force stop requested', severity: 'warn', source: WATCHER_SOURCE });
      // Try to start again with current config
      await StartWatching(config);
      isWatching = true;
      showWatcherConflictModal = false;
      conflictError = null;
      toast.success('Watcher restarted after force stop', { duration: 3000, position: 'top-right' });
    } catch (err) {
      console.error('Force-stop/start failed:', err);
      toast.error(`Force stop/start failed: ${err?.message || err}`, { duration: 5000, position: 'top-right' });
      pushLogEntry({ message: `Force stop/start failed: ${err?.message || err}`, severity: 'error', source: WATCHER_SOURCE, details: err });
    } finally {
      modalBusy = false;
    }
  }

  function cancelWatcherConflict() {
    showWatcherConflictModal = false;
    conflictError = null;
  }
</script>

<main>
  <Toaster />
  <div class="app-layout">
    <Sidebar
      {currentView}
      {isWatching}
      on:switchView={(e) => (currentView = e.detail)}
      on:openSettings={openSettings}
    />

    <!-- Main Content -->
    <div class="main-content">
      <!-- Error Boundary -->
      {#if hasError}
        <ErrorBoundary
          {errorMessage}
          on:retry={retryInitialization}
          on:dismiss={clearError}
        />
      {:else if isLoading}
        <LoadingBoundary />
      {:else}
        <!-- Configuration View -->
        {#if currentView === "configuration"}
          <ConfigurationView
            {config}
            {watchPathDisplay}
            {selectedPatternID}
            {availablePatterns}
            {availableNamers}
            {availableActions}
            {isWatching}
            {isBusy}
            on:selectDirectory={handleSelectDirectory}
            on:patternSelect={handlePatternSelect}
            on:selectActionDirectory={handleSelectActionDirectory}
            on:requestStart={handleStart}
            on:requestStop={handleStop}
            on:openAdvancedOperations={() => (currentView = "advancedOperations")}
            on:openMonitoring={() => (currentView = "monitoring")}
          />
        {:else if currentView === "monitoring"}
          <MonitoringView
            {stats}
            logs={monitoringLogs}
            {logContainer}
          />
        {:else if currentView === "advanced"}
          <!-- Advanced Options View -->
          <AdvancedView {config} {isWatching} on:updateConfig={handleUpdateConfig} />
        {:else if currentView === "advancedOperations"}
          <AdvancedOperationsView {config} on:error={handleAdvancedOperationsError} />
        {/if}
      {/if}
    </div>
  </div>

  <!-- Settings Modal -->
  {#if showSettingsModal}
    <SettingsModal
      {profiles}
      {selectedProfile}
      {isWatching}
      on:close={closeSettings}
      on:updateTheme={(e) => updateTheme(e.detail)}
      on:selectProfile={(e) => handleProfileSelect({ target: { value: e.detail } })}
      on:saveProfile={handleSaveProfile}
      on:deleteProfile={handleDeleteProfile}
      on:resetSettings={() => {
        settings.reset();
        toast("Settings reset to defaults", {
          duration: 3000,
          position: "top-right",
        });
      }}
    />
  {/if}

  <!-- Watcher conflict modal (shown when StartWatching reports watcher already running) -->
  {#if showWatcherConflictModal}
    <div class="modal-overlay" role="dialog" aria-modal="true">
      <div class="modal-card">
        <h3>Watcher already running</h3>
        <p>
          A watcher process is already running. You can attach to the running watcher (read-only),
          or force it to stop and restart. For safety, attaching is recommended.
        </p>
        {#if conflictError}
          <details class="conflict-details"><summary>Details</summary>
            <pre>{String(conflictError)}</pre>
          </details>
        {/if}
        <div class="modal-actions">
          <button class="btn" on:click={attachToRunningWatcher} disabled={modalBusy}>Attach</button>
          <button class="btn danger" on:click={forceStopAndStart} disabled={modalBusy}>Force Stop & Restart</button>
          <button class="btn" on:click={cancelWatcherConflict} disabled={modalBusy}>Cancel</button>
        </div>
      </div>
    </div>
  {/if}
</main>

<style>
  /* Global Variables - Optimized for Desktop */
  :root {
    --primary-bg: #0a0a0a;
    --secondary-bg: #1a1a2e;
    --card-bg: #161d31;
    --sidebar-bg: #0f0f23; /* Overridden by JS */
    --accent-primary: #8b5cf6;
    --accent-secondary: #ec4899;
    --text-primary: #ffffff;
    --text-secondary: #a1a1aa;
    --text-muted: #71717a;
    --border-color: #2a2a3a;
    --success-color: #10b981;
    --warning-color: #f59e0b;
    --error-color: #ef4444;
    --font-family: "Nunito", -apple-system, BlinkMacSystemFont, "Segoe UI",
      Roboto, sans-serif;
    --border-radius: 16px;
    --shadow: 0 8px 32px rgba(0, 0, 0, 0.2);
    --shadow-hover: 0 12px 40px rgba(139, 92, 246, 0.15);
  }

  /* Base Styles - Desktop Optimized */
  main {
    font-family: var(--font-family);
    background: linear-gradient(
      135deg,
      var(--primary-bg) 0%,
      var(--secondary-bg) 100%
    );
    color: var(--text-primary);
    height: 100vh;
    margin: 0;
    padding: 0;
    overflow: hidden;
    display: flex;
    flex-direction: column;
  }

  /* App Layout - Optimized Proportions */
  .app-layout {
    display: flex;
    height: 100vh;
    position: relative;
    overflow: hidden;
  }

  /* Main Content - Optimized Grid Layout */
  .main-content {
    flex: 1;
    display: grid;
    grid-template-rows: 1fr;
    overflow: hidden;
    background: linear-gradient(
      135deg,
      var(--primary-bg) 0%,
      var(--secondary-bg) 100%
    );
    min-height: 0; /* Ensure grid item can shrink */
  }

  /* Modal - watcher conflict */
  .modal-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0,0,0,0.6);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 60;
  }

  .modal-card {
    background: var(--card-bg);
    border: 1px solid var(--border-color);
    padding: 1.25rem;
    border-radius: 12px;
    width: min(600px, 92%);
    box-shadow: var(--shadow);
  }

  .modal-card h3 {
    margin: 0 0 0.5rem 0;
    font-size: 1.05rem;
  }

  .modal-actions {
    display: flex;
    gap: 0.5rem;
    justify-content: flex-end;
    margin-top: 1rem;
  }

  .modal-card pre {
    max-height: 140px;
    overflow: auto;
    background: rgba(255,255,255,0.02);
    padding: 0.5rem;
    border-radius: 6px;
    font-family: monospace;
    font-size: 0.8rem;
  }

  .btn {
    padding: 0.5rem 0.75rem;
    border-radius: 6px;
    background: var(--accent-primary);
    color: white;
    border: none;
    cursor: pointer;
  }

  .btn.danger {
    background: var(--error-color);
  }
</style>
