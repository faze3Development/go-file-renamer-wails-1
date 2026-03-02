import React, { useCallback } from 'react'
import { AppProvider, useApp } from './context/AppContext'
import type { Config, BackendVerbosity, AppSettings } from './context/AppContext'
import Sidebar from './components/Sidebar'
import ConfigurationView from './components/ConfigurationView'
import MonitoringView from './components/MonitoringView'
import SettingsModal from './components/SettingsModal'
import ErrorBoundary from './components/ErrorBoundary'
import LoadingBoundary from './components/LoadingBoundary'
import AdvancedView from './components/views/AdvancedView'
import AdvancedOperationsView from './components/views/AdvancedOperationsView'
import type { ErrorPayload } from './lib/errorBus'
import { fromErrorPayload } from './context/AppContext'

function WatcherConflictModal() {
  const { state, attachToRunningWatcher, forceStopAndStart, cancelWatcherConflict } = useApp()
  if (!state.showWatcherConflictModal) return null

  const errorMsg = state.conflictError instanceof Error
    ? state.conflictError.message
    : state.conflictError ? String(state.conflictError) : 'A watcher is already running.'

  return (
    <div className="modal-overlay" onClick={(e) => e.target === e.currentTarget && cancelWatcherConflict()}>
      <div className="modal-box" role="dialog" aria-modal="true" aria-labelledby="conflict-modal-title">
        <h3 id="conflict-modal-title">Watcher Already Running</h3>
        <p className="modal-message">{errorMsg}</p>
        <p className="modal-hint">You can attach to the running watcher, force stop and restart it, or cancel.</p>
        <div className="modal-actions">
          <button
            className="btn btn-primary"
            onClick={attachToRunningWatcher}
            disabled={state.modalBusy}
          >
            Attach to Running
          </button>
          <button
            className="btn btn-warning"
            onClick={forceStopAndStart}
            disabled={state.modalBusy}
          >
            {state.modalBusy ? 'Processing…' : 'Force Stop & Restart'}
          </button>
          <button
            className="btn btn-ghost"
            onClick={cancelWatcherConflict}
            disabled={state.modalBusy}
          >
            Cancel
          </button>
        </div>
      </div>
    </div>
  )
}

function AppShell() {
  const {
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
  } = useApp()

  const { config, isWatching, isBusy, currentView, showSettingsModal, settings, profiles, selectedProfile,
    availableNamers, availableActions, availablePatterns, selectedPatternID, watchPathDisplay,
    hasError, errorMessage, isLoading, stats, logs } = state

  const handleUpdateConfig = useCallback((key: keyof Config, value: unknown) => {
    dispatch({ type: 'UPDATE_CONFIG', payload: { key, value } })
  }, [dispatch])

  const handleAdvancedOperationsError = useCallback((payload: ErrorPayload) => {
    dispatch({ type: 'PUSH_LOG', payload: { message: `[${payload.context || 'advanced-ops'}] ${payload.message}`, severity: payload.severity || 'error', source: 'advanced-operations' } })
  }, [dispatch])

  const handleUpdateTheme = useCallback((theme: string) => {
    dispatch({ type: 'UPDATE_SETTING', payload: { key: 'theme', value: theme } })
  }, [dispatch])

  const handleUpdateSetting = useCallback((key: keyof AppSettings, value: unknown) => {
    dispatch({ type: 'UPDATE_SETTING', payload: { key, value } })
  }, [dispatch])

  const handleUpdateBackendVerbosity = useCallback((key: keyof BackendVerbosity, value: boolean) => {
    dispatch({ type: 'UPDATE_BACKEND_VERBOSITY', payload: { key, value } })
  }, [dispatch])

  const handleResetSettings = useCallback(() => {
    dispatch({ type: 'RESET_SETTINGS' })
  }, [dispatch])

  const handleSwitchView = useCallback((view: typeof currentView) => {
    dispatch({ type: 'SET_VIEW', payload: view })
  }, [dispatch])

  const handleOpenSettings = useCallback(() => {
    dispatch({ type: 'SET_SHOW_SETTINGS_MODAL', payload: true })
  }, [dispatch])

  const handleCloseSettings = useCallback(() => {
    dispatch({ type: 'SET_SHOW_SETTINGS_MODAL', payload: false })
  }, [dispatch])

  if (isLoading) {
    return <LoadingBoundary />
  }

  if (hasError) {
    return (
      <ErrorBoundary
        errorMessage={errorMessage}
        onRetry={retryInitialization}
        onDismiss={() => dispatch({ type: 'CLEAR_ERROR' })}
      />
    )
  }

  return (
    <div className={`app-container${settings.compactMode ? ' compact' : ''}`}>
      <Sidebar
        currentView={currentView}
        isWatching={isWatching}
        onSwitchView={handleSwitchView}
        onOpenSettings={handleOpenSettings}
      />

      <main className="main-content">
        {currentView === 'configuration' && (
          <ConfigurationView
            config={config}
            watchPathDisplay={watchPathDisplay}
            selectedPatternID={selectedPatternID}
            availablePatterns={availablePatterns}
            availableNamers={availableNamers}
            availableActions={availableActions}
            isWatching={isWatching}
            isBusy={isBusy}
            onSelectDirectory={handleSelectDirectory}
            onPatternSelect={handlePatternSelect}
            onUpdateConfig={handleUpdateConfig}
            onSelectActionDirectory={handleSelectActionDirectory}
            onRequestStart={handleStart}
            onRequestStop={handleStop}
            onOpenAdvancedOperations={() => handleSwitchView('advancedOperations')}
            onOpenMonitoring={() => handleSwitchView('monitoring')}
          />
        )}

        {currentView === 'monitoring' && (
          <MonitoringView
            stats={stats}
            logs={logs}
          />
        )}

        {currentView === 'advanced' && (
          <AdvancedView
            config={config}
            isWatching={isWatching}
            onUpdateConfig={handleUpdateConfig}
          />
        )}

        {currentView === 'advancedOperations' && (
          <AdvancedOperationsView
            config={config}
            onError={handleAdvancedOperationsError}
          />
        )}
      </main>

      {showSettingsModal && (
        <SettingsModal
          settings={settings}
          profiles={profiles}
          selectedProfile={selectedProfile}
          isWatching={isWatching}
          onClose={handleCloseSettings}
          onUpdateTheme={handleUpdateTheme}
          onUpdateSetting={handleUpdateSetting}
          onUpdateBackendVerbosity={handleUpdateBackendVerbosity}
          onResetSettings={handleResetSettings}
          onSelectProfile={handleProfileSelect}
          onSaveProfile={handleSaveProfile}
          onDeleteProfile={handleDeleteProfile}
        />
      )}

      <WatcherConflictModal />
    </div>
  )
}

export default function App() {
  return (
    <AppProvider>
      <AppShell />
    </AppProvider>
  )
}
