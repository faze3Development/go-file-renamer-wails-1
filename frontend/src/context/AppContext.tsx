import React, { createContext, useContext, useReducer, useEffect, useCallback, useRef } from 'react'
import toast from 'react-hot-toast'
import { EventsOn, EventsOff, OnFileDrop, OnFileDropOff } from '../../wailsjs/runtime/runtime.js'
import {
  LoadProfiles, GetNamerInfo, GetActionInfo, GetPatternInfo,
  SaveProfile, DeleteProfile, SelectActionDirectory, SelectDirectory,
  StartWatching, StopWatching
} from '../../wailsjs/go/main/App.js'
import { appendLog, fromErrorPayload, LOG_STORE_DEFAULT_LIMIT, LOG_STORE_HARD_LIMIT } from '../lib/logStore'
import type { LogEntry, LogEntryInput } from '../lib/logStore'
import { isLogEntryPayload, isStatsPayload } from '../types/events'
import type { StatsPayload, LogEntryPayload } from '../types/events'

// ─── Theme definitions ────────────────────────────────────────────────────────

export interface ThemeColors {
  '--primary-bg': string
  '--secondary-bg': string
  '--card-bg': string
  '--sidebar-bg': string
  '--accent-primary': string
  '--accent-secondary': string
  '--text-primary': string
  '--text-secondary': string
  '--text-muted': string
  '--border-color': string
  '--success-color': string
  '--warning-color': string
  '--error-color': string
}

export interface Theme {
  name: string
  description: string
  colors: ThemeColors
}

export const themes: Record<string, Theme> = {
  default: {
    name: 'Starlight',
    description: 'Deep space with vibrant purple and blue nebulas.',
    colors: {
      '--primary-bg': '#0D1117',
      '--secondary-bg': '#161B22',
      '--card-bg': '#0D1117',
      '--sidebar-bg': '#010409',
      '--accent-primary': '#8b5cf6',
      '--accent-secondary': '#3b82f6',
      '--text-primary': '#e6edf3',
      '--text-secondary': '#7d8590',
      '--text-muted': '#586069',
      '--border-color': '#30363d',
      '--success-color': '#238636',
      '--warning-color': '#d29922',
      '--error-color': '#f85149',
    },
  },
  cyberpunk: {
    name: 'Cyberpunk',
    description: 'High-tech, low-life. Neon yellow and cyan.',
    colors: {
      '--primary-bg': '#000000',
      '--secondary-bg': '#0c0c0c',
      '--card-bg': '#050505',
      '--sidebar-bg': '#000000',
      '--accent-primary': '#fcee0a',
      '--accent-secondary': '#00f0ff',
      '--text-primary': '#ffffff',
      '--text-secondary': '#b0b0b0',
      '--text-muted': '#6a6a6a',
      '--border-color': '#333333',
      '--success-color': '#00ff7f',
      '--warning-color': '#ffae00',
      '--error-color': '#ff003c',
    },
  },
  forest: {
    name: 'Forest',
    description: 'Earthy greens and warm, natural tones.',
    colors: {
      '--primary-bg': '#1a201a',
      '--secondary-bg': '#202820',
      '--card-bg': '#1a201a',
      '--sidebar-bg': '#101410',
      '--accent-primary': '#4ade80',
      '--accent-secondary': '#f97316',
      '--text-primary': '#f0fdf4',
      '--text-secondary': '#a3a3a3',
      '--text-muted': '#6b7280',
      '--border-color': '#374151',
      '--success-color': '#10b981',
      '--warning-color': '#f59e0b',
      '--error-color': '#ef4444',
    },
  },
}

export function applyTheme(themeName: string): void {
  const theme = themes[themeName]
  if (!theme) {
    if (themeName !== 'default') applyTheme('default')
    return
  }
  const root = document.documentElement
  Object.entries(theme.colors).forEach(([prop, val]) => {
    root.style.setProperty(prop, val)
  })
}

// ─── Settings ─────────────────────────────────────────────────────────────────

export interface BackendVerbosity {
  global: boolean
  watcher: boolean
  advancedOperations: boolean
}

export interface AppSettings {
  theme: string
  showActivityLog: boolean
  showStats: boolean
  compactMode: boolean
  logRetentionLimit: number
  backendVerbosity: BackendVerbosity
}

export const LOG_RETENTION_DEFAULT = 1000
export const LOG_RETENTION_SOFT_MAX = 5000

export const BACKEND_VERBOSITY_DEFAULT: BackendVerbosity = {
  global: false,
  watcher: true,
  advancedOperations: true,
}

const SETTINGS_KEY = 'fileRenamerSettings'

function sanitizeLogRetention(value: unknown): number {
  const numeric = Number(value)
  if (!Number.isFinite(numeric) || numeric <= 0) return LOG_RETENTION_DEFAULT
  return Math.min(Math.max(Math.floor(numeric), 100), LOG_STORE_HARD_LIMIT)
}

function sanitizeBackendVerbosity(value: unknown): BackendVerbosity {
  const sanitized = { ...BACKEND_VERBOSITY_DEFAULT }
  if (value && typeof value === 'object') {
    ;(Object.keys(sanitized) as (keyof BackendVerbosity)[]).forEach((key) => {
      if (key in (value as object)) {
        sanitized[key] = Boolean((value as Record<string, unknown>)[key])
      }
    })
  }
  return sanitized
}

function loadSettings(): AppSettings {
  const defaults: AppSettings = {
    theme: 'default',
    showActivityLog: true,
    showStats: true,
    compactMode: false,
    logRetentionLimit: LOG_RETENTION_DEFAULT,
    backendVerbosity: { ...BACKEND_VERBOSITY_DEFAULT },
  }
  try {
    const stored = localStorage.getItem(SETTINGS_KEY)
    if (!stored) return defaults
    const parsed = JSON.parse(stored)
    return {
      ...defaults,
      ...parsed,
      logRetentionLimit: sanitizeLogRetention(parsed.logRetentionLimit),
      backendVerbosity: sanitizeBackendVerbosity(parsed.backendVerbosity),
    }
  } catch {
    return defaults
  }
}

function persistSettings(settings: AppSettings): void {
  localStorage.setItem(SETTINGS_KEY, JSON.stringify(settings))
}

// ─── Config type ──────────────────────────────────────────────────────────────

export interface Config {
  WatchPaths: string[]
  Recursive: boolean
  DryRun: boolean
  NamePattern: string
  TemplateString: string
  DateTimeFormat: string
  RandomLength: number
  Settle: number
  SettleTimeout: number
  Retries: number
  NoInitialScan: boolean
  NamerID: string
  ActionID: string
  ActionConfig: Record<string, string>
}

// ─── Info types ───────────────────────────────────────────────────────────────

export interface NamerInfo { id: string; name: string; description: string }
export interface ActionInfo { id: string; name: string; description: string }
export interface PatternInfo { id: string; name: string; regex: string; description: string }

// ─── App State ────────────────────────────────────────────────────────────────

export type ViewType = 'configuration' | 'monitoring' | 'advanced' | 'advancedOperations'

export interface AppState {
  config: Config
  watchPathDisplay: string
  isWatching: boolean
  isBusy: boolean
  showWatcherConflictModal: boolean
  conflictError: unknown
  modalBusy: boolean
  stats: StatsPayload
  logs: LogEntry[]
  currentView: ViewType
  showSettingsModal: boolean
  settings: AppSettings
  profiles: Record<string, Config>
  selectedProfile: string
  availableNamers: NamerInfo[]
  availableActions: ActionInfo[]
  availablePatterns: PatternInfo[]
  selectedPatternID: string
  hasError: boolean
  errorMessage: string
  isLoading: boolean
  isInitialized: boolean
}

const DEFAULT_CONFIG: Config = {
  WatchPaths: [],
  Recursive: true,
  DryRun: false,
  NamePattern: '(?i)^(?:untitled|screenshot(?:[\\s_-]\\d{4}-\\d{2}-\\d{2})?|untitiled)(?:[\\s_-]*(?:\\(\\d+\\)|\\d+))?$',
  TemplateString: '{original}-{date}',
  DateTimeFormat: '2006-01-02_15-04-05',
  RandomLength: 12,
  Settle: 750,
  SettleTimeout: 10,
  Retries: 6,
  NoInitialScan: false,
  NamerID: 'random',
  ActionID: 'none',
  ActionConfig: { destinationPath: '' },
}

// ─── Reducer ──────────────────────────────────────────────────────────────────

export type AppAction =
  | { type: 'SET_WATCHING'; payload: boolean }
  | { type: 'SET_BUSY'; payload: boolean }
  | { type: 'SET_STATS'; payload: StatsPayload }
  | { type: 'PUSH_LOG'; payload: LogEntryInput | LogEntryInput[] }
  | { type: 'CLEAR_LOGS' }
  | { type: 'SET_VIEW'; payload: ViewType }
  | { type: 'SET_SHOW_SETTINGS_MODAL'; payload: boolean }
  | { type: 'SET_CONFIG'; payload: Config }
  | { type: 'UPDATE_CONFIG'; payload: { key: keyof Config; value: unknown } }
  | { type: 'SET_WATCH_PATH'; payload: string }
  | { type: 'SET_PROFILES'; payload: Record<string, Config> }
  | { type: 'SET_SELECTED_PROFILE'; payload: string }
  | { type: 'SET_NAMERS'; payload: NamerInfo[] }
  | { type: 'SET_ACTIONS'; payload: ActionInfo[] }
  | { type: 'SET_PATTERNS'; payload: PatternInfo[] }
  | { type: 'SET_SELECTED_PATTERN_ID'; payload: string }
  | { type: 'SET_ERROR'; payload: { message: string } }
  | { type: 'CLEAR_ERROR' }
  | { type: 'SET_LOADING'; payload: boolean }
  | { type: 'SET_INITIALIZED'; payload: boolean }
  | { type: 'SET_WATCHER_CONFLICT'; payload: { show: boolean; error: unknown } }
  | { type: 'SET_MODAL_BUSY'; payload: boolean }
  | { type: 'UPDATE_SETTING'; payload: { key: keyof AppSettings; value: unknown } }
  | { type: 'UPDATE_BACKEND_VERBOSITY'; payload: { key: keyof BackendVerbosity; value: boolean } }
  | { type: 'RESET_SETTINGS' }

function appReducer(state: AppState, action: AppAction): AppState {
  switch (action.type) {
    case 'SET_WATCHING':
      return { ...state, isWatching: action.payload }
    case 'SET_BUSY':
      return { ...state, isBusy: action.payload }
    case 'SET_STATS':
      return { ...state, stats: { ...state.stats, ...action.payload } }
    case 'PUSH_LOG':
      return { ...state, logs: appendLog(state.logs, action.payload, state.settings.logRetentionLimit) }
    case 'CLEAR_LOGS':
      return { ...state, logs: [] }
    case 'SET_VIEW':
      return { ...state, currentView: action.payload }
    case 'SET_SHOW_SETTINGS_MODAL':
      return { ...state, showSettingsModal: action.payload }
    case 'SET_CONFIG':
      return { ...state, config: action.payload }
    case 'UPDATE_CONFIG':
      return { ...state, config: { ...state.config, [action.payload.key]: action.payload.value } }
    case 'SET_WATCH_PATH':
      return { ...state, watchPathDisplay: action.payload }
    case 'SET_PROFILES':
      return { ...state, profiles: action.payload }
    case 'SET_SELECTED_PROFILE':
      return { ...state, selectedProfile: action.payload }
    case 'SET_NAMERS':
      return { ...state, availableNamers: action.payload }
    case 'SET_ACTIONS':
      return { ...state, availableActions: action.payload }
    case 'SET_PATTERNS':
      return { ...state, availablePatterns: action.payload }
    case 'SET_SELECTED_PATTERN_ID':
      return { ...state, selectedPatternID: action.payload }
    case 'SET_ERROR':
      return { ...state, hasError: true, errorMessage: action.payload.message }
    case 'CLEAR_ERROR':
      return { ...state, hasError: false, errorMessage: '' }
    case 'SET_LOADING':
      return { ...state, isLoading: action.payload }
    case 'SET_INITIALIZED':
      return { ...state, isInitialized: action.payload }
    case 'SET_WATCHER_CONFLICT':
      return { ...state, showWatcherConflictModal: action.payload.show, conflictError: action.payload.error }
    case 'SET_MODAL_BUSY':
      return { ...state, modalBusy: action.payload }
    case 'UPDATE_SETTING': {
      let nextValue: unknown = action.payload.value
      if (action.payload.key === 'logRetentionLimit') nextValue = sanitizeLogRetention(action.payload.value)
      else if (action.payload.key === 'backendVerbosity') nextValue = sanitizeBackendVerbosity(action.payload.value)
      const newSettings: AppSettings = { ...state.settings, [action.payload.key]: nextValue }
      if (action.payload.key === 'theme') applyTheme(nextValue as string)
      persistSettings(newSettings)
      return { ...state, settings: newSettings }
    }
    case 'UPDATE_BACKEND_VERBOSITY': {
      const nextVerbosity: BackendVerbosity = {
        ...state.settings.backendVerbosity,
        [action.payload.key]: action.payload.value,
      }
      const newSettings: AppSettings = { ...state.settings, backendVerbosity: nextVerbosity }
      persistSettings(newSettings)
      return { ...state, settings: newSettings }
    }
    case 'RESET_SETTINGS': {
      const resetSettings: AppSettings = {
        theme: 'default',
        showActivityLog: true,
        showStats: true,
        compactMode: false,
        logRetentionLimit: LOG_RETENTION_DEFAULT,
        backendVerbosity: { ...BACKEND_VERBOSITY_DEFAULT },
      }
      applyTheme('default')
      persistSettings(resetSettings)
      return { ...state, settings: resetSettings }
    }
    default:
      return state
  }
}

// ─── Context ──────────────────────────────────────────────────────────────────

interface AppContextValue {
  state: AppState
  dispatch: React.Dispatch<AppAction>
  handleStart: () => Promise<void>
  handleStop: () => Promise<void>
  handleSelectDirectory: () => Promise<void>
  handleSelectActionDirectory: () => Promise<void>
  handlePatternSelect: (id: string) => void
  handleProfileSelect: (profileName: string) => void
  handleSaveProfile: () => Promise<void>
  handleDeleteProfile: () => Promise<void>
  retryInitialization: () => void
  attachToRunningWatcher: () => void
  forceStopAndStart: () => Promise<void>
  cancelWatcherConflict: () => void
  loadInitialData: () => Promise<boolean>
}

const AppContext = createContext<AppContextValue | null>(null)

// ─── WATCHER_SOURCE constant ──────────────────────────────────────────────────

const WATCHER_SOURCE = 'watcher'
const ADVANCED_OPERATIONS_SOURCE = 'advanced-operations'

const severityRank: Record<string, number> = {
  debug: 0, info: 1, warn: 2, error: 3, fatal: 4,
}

function withTimeout<T>(promise: Promise<T>, ms = 5000): Promise<T> {
  return new Promise((resolve, reject) => {
    const timer = setTimeout(() => reject(new Error('Request timed out')), ms)
    promise.then(resolve, reject).finally(() => clearTimeout(timer))
  })
}

// ─── Provider ─────────────────────────────────────────────────────────────────

export function AppProvider({ children }: { children: React.ReactNode }) {
  const initialSettings = loadSettings()
  const initialState: AppState = {
    config: { ...DEFAULT_CONFIG },
    watchPathDisplay: '',
    isWatching: false,
    isBusy: false,
    showWatcherConflictModal: false,
    conflictError: null,
    modalBusy: false,
    stats: { scanned: 0, renamed: 0, skipped: 0, errors: 0 },
    logs: [],
    currentView: 'configuration',
    showSettingsModal: false,
    settings: initialSettings,
    profiles: {},
    selectedProfile: '',
    availableNamers: [],
    availableActions: [],
    availablePatterns: [],
    selectedPatternID: 'default_untitled',
    hasError: false,
    errorMessage: '',
    isLoading: true,
    isInitialized: false,
  }

  const [state, dispatch] = useReducer(appReducer, initialState)

  // Keep a ref to state so event handlers always have the latest values
  const stateRef = useRef(state)
  useEffect(() => { stateRef.current = state }, [state])

  // Apply initial theme on mount
  useEffect(() => {
    applyTheme(initialSettings.theme)
  }, []) // eslint-disable-line

  // ─── Log ingestion ──────────────────────────────────────────────────────────

  const ingestBackendLog = useCallback((payload: unknown) => {
    if (!isLogEntryPayload(payload)) return
    const entry: LogEntryPayload = {
      message: payload.message,
      severity: payload.severity ?? 'info',
      source: payload.source ?? 'backend',
      context: payload.context,
      details: payload.details,
      metadata: payload.metadata,
      timestamp: payload.timestamp ?? new Date().toISOString(),
      id: payload.id,
    }

    const currentSettings = stateRef.current.settings
    const backendVerbosityPrefs = currentSettings.backendVerbosity
    const rank = severityRank[entry.severity ?? 'info'] ?? severityRank.info
    if (rank < severityRank.warn) {
      let prefKey: keyof BackendVerbosity
      switch (entry.source) {
        case WATCHER_SOURCE: prefKey = 'watcher'; break
        case ADVANCED_OPERATIONS_SOURCE: prefKey = 'advancedOperations'; break
        default: prefKey = 'global'
      }
      if (!backendVerbosityPrefs[prefKey]) return
      if (entry.severity === 'debug' && !backendVerbosityPrefs.global) return
    }

    dispatch({ type: 'PUSH_LOG', payload: entry })
  }, [])

  // ─── Wails event listeners ──────────────────────────────────────────────────

  const lastWatcherStoppedAtRef = useRef(0)

  useEffect(() => {
    const initListeners = () => {
      if (typeof window === 'undefined' || !(window as Window & { runtime?: unknown }).runtime) {
        setTimeout(initListeners, 100)
        return
      }

      EventsOn('log:entry', (payload: unknown) => {
        try { ingestBackendLog(payload) } catch (e) { console.warn('log:entry handler error:', e) }
      })

      // Legacy event alias
      EventsOn('logEntry', (payload: unknown) => {
        try { ingestBackendLog(payload) } catch (e) { console.warn('logEntry handler error:', e) }
      })

      EventsOn('watcherStarted', () => {
        dispatch({ type: 'SET_WATCHING', payload: true })
        dispatch({ type: 'SET_BUSY', payload: false })
      })

      EventsOn('watcherStopped', () => {
        const now = Date.now()
        if (now - lastWatcherStoppedAtRef.current < 2000) return
        lastWatcherStoppedAtRef.current = now
        dispatch({ type: 'SET_WATCHING', payload: false })
        dispatch({ type: 'SET_BUSY', payload: false })
        toast('File watching stopped', { duration: 3000 })
      })

      EventsOn('statsUpdated', (newStats: unknown) => {
        if (isStatsPayload(newStats)) dispatch({ type: 'SET_STATS', payload: newStats })
      })

      OnFileDrop((_x: number, _y: number, paths: string[]) => {
        if (stateRef.current.isWatching) return
        if (paths?.length > 0) {
          dispatch({ type: 'SET_WATCH_PATH', payload: paths[0] })
          dispatch({ type: 'UPDATE_CONFIG', payload: { key: 'WatchPaths', value: [paths[0]] } })
        }
      }, false)
    }

    initListeners()

    return () => {
      try {
        EventsOff('log:entry', 'logEntry', 'watcherStarted', 'watcherStopped', 'statsUpdated')
      } catch (e) { console.warn('EventsOff error:', e) }
      try { OnFileDropOff() } catch (e) { console.warn('OnFileDropOff error:', e) }
    }
  }, [ingestBackendLog])

  // ─── Initial data load ──────────────────────────────────────────────────────

  const loadInitialData = useCallback(async (): Promise<boolean> => {
    try {
      dispatch({ type: 'SET_LOADING', payload: true })
      dispatch({ type: 'CLEAR_ERROR' })

      const win = window as Window & { runtime?: unknown; go?: unknown }
      if (win.runtime && win.go) {
        const [loadedProfiles, loadedNamers, loadedActions, loadedPatterns] = await Promise.allSettled([
          withTimeout(LoadProfiles()),
          withTimeout(GetNamerInfo()),
          withTimeout(GetActionInfo()),
          withTimeout(GetPatternInfo()),
        ])

        const profiles = loadedProfiles.status === 'fulfilled' ? (loadedProfiles.value as Record<string, Config>) || {} : {}
        const namers = loadedNamers.status === 'fulfilled' ? (loadedNamers.value as NamerInfo[]) || [] : []
        const actions = loadedActions.status === 'fulfilled' ? (loadedActions.value as ActionInfo[]) || [] : []
        const patterns = loadedPatterns.status === 'fulfilled' ? (loadedPatterns.value as PatternInfo[]) || [] : []

        dispatch({ type: 'SET_PROFILES', payload: profiles })
        dispatch({ type: 'SET_NAMERS', payload: namers })
        dispatch({ type: 'SET_ACTIONS', payload: actions })
        dispatch({ type: 'SET_PATTERNS', payload: patterns })

        if (patterns.length > 0) {
          const currentConfig = stateRef.current.config
          const match = patterns.find((p) => p.regex === currentConfig.NamePattern)
          if (match) {
            dispatch({ type: 'SET_SELECTED_PATTERN_ID', payload: match.id })
          } else {
            dispatch({ type: 'SET_SELECTED_PATTERN_ID', payload: patterns[0].id })
            dispatch({ type: 'UPDATE_CONFIG', payload: { key: 'NamePattern', value: patterns[0].regex } })
          }
        }
      } else {
        // Fallback for browser dev mode
        dispatch({ type: 'SET_PATTERNS', payload: [{ id: 'default_untitled', name: 'Default Untitled/Screenshot', regex: '(?i)^(?:untitled|screenshot(?:[\\s_-]\\d{4}-\\d{2}-\\d{2})?|untitled)(?:[\\s_-]*(?:\\(\\d+\\)|\\d+))?$', description: 'Matches untitled files and screenshots' }] })
        dispatch({ type: 'SET_NAMERS', payload: [{ id: 'random', name: 'Random Name', description: 'Generate random alphanumeric names' }, { id: 'template', name: 'Template Based', description: 'Use template with placeholders' }, { id: 'timestamp', name: 'Timestamp', description: 'Use current timestamp as filename' }] })
        dispatch({ type: 'SET_ACTIONS', payload: [{ id: 'none', name: 'No Action', description: 'Just rename the file in place' }, { id: 'move', name: 'Move File', description: 'Move file to another directory' }, { id: 'copy', name: 'Copy File', description: 'Copy file to another directory' }, { id: 'advanced_operations', name: 'Advanced File Operations', description: 'Run the advanced processing pipeline' }] })
        dispatch({ type: 'SET_PROFILES', payload: {} })
        dispatch({ type: 'SET_SELECTED_PATTERN_ID', payload: 'default_untitled' })
      }

      dispatch({ type: 'SET_INITIALIZED', payload: true })
      return true
    } catch (error) {
      const msg = (error instanceof Error ? error.message : String(error))
      dispatch({ type: 'SET_ERROR', payload: { message: msg } })
      return false
    } finally {
      dispatch({ type: 'SET_LOADING', payload: false })
    }
  }, [])

  // Run initial load on mount
  useEffect(() => { loadInitialData() }, [loadInitialData])

  // ─── Action handlers ────────────────────────────────────────────────────────

  const checkRuntime = () => {
    const win = window as Window & { runtime?: unknown }
    if (!win.runtime) {
      toast('Runtime not available - please wait for app to fully load', { duration: 4000 })
      return false
    }
    return true
  }

  const handleStart = useCallback(async () => {
    const { watchPathDisplay, config } = stateRef.current
    if (!watchPathDisplay) {
      toast('Please select a directory to watch first.', { duration: 3000 })
      return
    }
    if (!checkRuntime()) return

    dispatch({ type: 'SET_BUSY', payload: true })
    dispatch({ type: 'CLEAR_LOGS' })
    dispatch({ type: 'SET_STATS', payload: { scanned: 0, renamed: 0, skipped: 0, errors: 0 } })

    try {
      await StartWatching(config)
      dispatch({ type: 'SET_WATCHING', payload: true })
      toast.success('File watching started successfully', { duration: 3000 })
    } catch (error) {
      const msg = (error instanceof Error ? error.message : String(error))
      if (msg.toLowerCase().includes('already running')) {
        dispatch({ type: 'SET_WATCHER_CONFLICT', payload: { show: true, error } })
        dispatch({ type: 'PUSH_LOG', payload: { message: `Watcher already running: ${msg}`, severity: 'warn', source: WATCHER_SOURCE } })
        toast(`Watcher already running`, { duration: 5000 })
      } else {
        toast.error(`Failed to start watching: ${msg}`, { duration: 5000 })
        dispatch({ type: 'PUSH_LOG', payload: { message: `Failed to start watcher: ${msg}`, severity: 'error', source: WATCHER_SOURCE } })
      }
      dispatch({ type: 'SET_BUSY', payload: false })
    }
  }, [])

  const handleStop = useCallback(async () => {
    if (!checkRuntime()) return
    dispatch({ type: 'SET_BUSY', payload: true })
    try {
      await StopWatching()
      dispatch({ type: 'SET_WATCHING', payload: false })
      dispatch({ type: 'PUSH_LOG', payload: { message: 'Watcher stop requested', severity: 'info', source: WATCHER_SOURCE } })
    } catch (error) {
      const msg = (error instanceof Error ? error.message : String(error))
      toast.error(`Failed to stop watching: ${msg}`, { duration: 5000 })
      dispatch({ type: 'SET_BUSY', payload: false })
      dispatch({ type: 'PUSH_LOG', payload: { message: `Failed to stop watcher: ${msg}`, severity: 'error', source: WATCHER_SOURCE } })
    }
  }, [])

  const handleSelectDirectory = useCallback(async () => {
    if (!checkRuntime()) return
    try {
      const selectedPath = await SelectDirectory()
      if (selectedPath) {
        dispatch({ type: 'SET_WATCH_PATH', payload: selectedPath })
        dispatch({ type: 'UPDATE_CONFIG', payload: { key: 'WatchPaths', value: [selectedPath] } })
      }
    } catch (error) {
      const msg = (error instanceof Error ? error.message : String(error))
      toast.error(`Failed to select directory: ${msg}`, { duration: 5000 })
    }
  }, [])

  const handleSelectActionDirectory = useCallback(async () => {
    if (!checkRuntime()) return
    try {
      const selectedPath = await SelectActionDirectory()
      if (selectedPath) {
        const currentActionConfig = stateRef.current.config.ActionConfig || {}
        dispatch({ type: 'UPDATE_CONFIG', payload: { key: 'ActionConfig', value: { ...currentActionConfig, destinationPath: selectedPath } } })
      }
    } catch (error) {
      const msg = (error instanceof Error ? error.message : String(error))
      toast.error(`Failed to select directory: ${msg}`, { duration: 5000 })
    }
  }, [])

  const handlePatternSelect = useCallback((id: string) => {
    dispatch({ type: 'SET_SELECTED_PATTERN_ID', payload: id })
    if (id !== 'custom') {
      const pattern = stateRef.current.availablePatterns.find((p) => p.id === id)
      if (pattern?.regex) {
        dispatch({ type: 'UPDATE_CONFIG', payload: { key: 'NamePattern', value: pattern.regex } })
      }
    }
  }, [])

  const handleProfileSelect = useCallback((profileName: string) => {
    dispatch({ type: 'SET_SELECTED_PROFILE', payload: profileName })
    const profile = stateRef.current.profiles[profileName]
    if (profile && typeof profile === 'object') {
      const newConfig: Config = { ...stateRef.current.config, ...profile, ActionConfig: { ...stateRef.current.config.ActionConfig, ...(profile.ActionConfig || {}) } }
      dispatch({ type: 'SET_CONFIG', payload: newConfig })
      dispatch({ type: 'SET_WATCH_PATH', payload: newConfig.WatchPaths?.[0] || '' })

      const matchingPattern = stateRef.current.availablePatterns.find((p) => p.regex === newConfig.NamePattern)
      dispatch({ type: 'SET_SELECTED_PATTERN_ID', payload: matchingPattern ? matchingPattern.id : 'custom' })

      toast.success(`Profile '${profileName}' loaded successfully`, { duration: 3000 })
    }
  }, [])

  const handleSaveProfile = useCallback(async () => {
    if (!checkRuntime()) return
    const name = window.prompt('Enter a name for this profile:', stateRef.current.selectedProfile || 'New Profile')
    if (!name) return
    try {
      await SaveProfile(name, stateRef.current.config)
      await loadInitialData()
      dispatch({ type: 'SET_SELECTED_PROFILE', payload: name })
      toast.success(`Profile '${name}' saved successfully`, { duration: 3000 })
    } catch (error) {
      const msg = (error instanceof Error ? error.message : String(error))
      toast.error(`Failed to save profile: ${msg}`, { duration: 5000 })
    }
  }, [loadInitialData])

  const handleDeleteProfile = useCallback(async () => {
    const { selectedProfile } = stateRef.current
    if (!selectedProfile) {
      toast('No profile selected to delete', { duration: 3000 })
      return
    }
    if (!checkRuntime()) return
    if (!window.confirm(`Are you sure you want to delete the profile '${selectedProfile}'?`)) return
    try {
      await DeleteProfile(selectedProfile)
      dispatch({ type: 'SET_SELECTED_PROFILE', payload: '' })
      await loadInitialData()
      toast.success('Profile deleted successfully', { duration: 3000 })
    } catch (error) {
      const msg = (error instanceof Error ? error.message : String(error))
      toast.error(`Failed to delete profile: ${msg}`, { duration: 5000 })
    }
  }, [loadInitialData])

  const retryInitialization = useCallback(() => {
    dispatch({ type: 'SET_INITIALIZED', payload: false })
    dispatch({ type: 'CLEAR_ERROR' })
    loadInitialData()
  }, [loadInitialData])

  const attachToRunningWatcher = useCallback(() => {
    dispatch({ type: 'SET_WATCHER_CONFLICT', payload: { show: false, error: null } })
    dispatch({ type: 'SET_WATCHING', payload: true })
    dispatch({ type: 'PUSH_LOG', payload: { message: 'Attached to running watcher (UI only)', severity: 'info', source: WATCHER_SOURCE } })
    toast.success('Attached to running watcher', { duration: 3000 })
  }, [])

  const forceStopAndStart = useCallback(async () => {
    dispatch({ type: 'SET_MODAL_BUSY', payload: true })
    try {
      await StopWatching()
      dispatch({ type: 'PUSH_LOG', payload: { message: 'Force stop requested', severity: 'warn', source: WATCHER_SOURCE } })
      await StartWatching(stateRef.current.config)
      dispatch({ type: 'SET_WATCHING', payload: true })
      dispatch({ type: 'SET_WATCHER_CONFLICT', payload: { show: false, error: null } })
      toast.success('Watcher restarted after force stop', { duration: 3000 })
    } catch (error) {
      const msg = (error instanceof Error ? error.message : String(error))
      toast.error(`Force stop/start failed: ${msg}`, { duration: 5000 })
      dispatch({ type: 'PUSH_LOG', payload: { message: `Force stop/start failed: ${msg}`, severity: 'error', source: WATCHER_SOURCE } })
    } finally {
      dispatch({ type: 'SET_MODAL_BUSY', payload: false })
    }
  }, [])

  const cancelWatcherConflict = useCallback(() => {
    dispatch({ type: 'SET_WATCHER_CONFLICT', payload: { show: false, error: null } })
  }, [])

  const value: AppContextValue = {
    state,
    dispatch,
    handleStart,
    handleStop,
    handleSelectDirectory,
    handleSelectActionDirectory,
    handlePatternSelect,
    handleProfileSelect,
    handleSaveProfile,
    handleDeleteProfile,
    retryInitialization,
    attachToRunningWatcher,
    forceStopAndStart,
    cancelWatcherConflict,
    loadInitialData,
  }

  return <AppContext.Provider value={value}>{children}</AppContext.Provider>
}

export function useApp(): AppContextValue {
  const ctx = useContext(AppContext)
  if (!ctx) throw new Error('useApp must be used within AppProvider')
  return ctx
}

export { fromErrorPayload }
export type { LogEntry, LogEntryInput }
