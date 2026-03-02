import React, { useRef, useEffect } from 'react'
import { X, Palette, UserSquare, Layout, Sparkles, Terminal, Leaf, Save, Trash2 } from 'lucide-react'
import Toggle from './shared/Toggle'
import {
  themes, LOG_RETENTION_DEFAULT, LOG_RETENTION_SOFT_MAX,
} from '../context/AppContext'
import type { AppSettings, BackendVerbosity } from '../context/AppContext'
import { LOG_STORE_HARD_LIMIT as HARD_MAX } from '../lib/logStore'

interface Props {
  settings: AppSettings
  profiles: Record<string, unknown>
  selectedProfile: string
  isWatching: boolean
  onClose: () => void
  onUpdateTheme: (theme: string) => void
  onUpdateSetting: (key: keyof AppSettings, value: unknown) => void
  onUpdateBackendVerbosity: (key: keyof BackendVerbosity, value: boolean) => void
  onResetSettings: () => void
  onSelectProfile: (name: string) => void
  onSaveProfile: () => void
  onDeleteProfile: () => void
}

const verbosityOptions: { key: keyof BackendVerbosity; title: string; description: string }[] = [
  { key: 'global', title: 'Verbose Backend Events', description: 'Forward debug-level telemetry from all backend modules.' },
  { key: 'watcher', title: 'Watcher Diagnostics', description: 'Include filesystem watcher state changes and errors.' },
  { key: 'advancedOperations', title: 'Advanced Operations Diagnostics', description: 'Include rename pipeline insights and advanced ops traces.' },
]

export default function SettingsModal({
  settings, profiles, selectedProfile, isWatching,
  onClose, onUpdateTheme, onUpdateSetting, onUpdateBackendVerbosity, onResetSettings,
  onSelectProfile, onSaveProfile, onDeleteProfile,
}: Props) {
  const modalRef = useRef<HTMLDivElement>(null)

  useEffect(() => { modalRef.current?.focus() }, [])

  const retentionLimit = settings.logRetentionLimit ?? LOG_RETENTION_DEFAULT
  const retentionWarning = retentionLimit > LOG_RETENTION_SOFT_MAX

  const handleOverlayClick = (e: React.MouseEvent) => {
    if (e.target === e.currentTarget) onClose()
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Escape') onClose()
  }

  const themeIcons: Record<string, React.ReactNode> = {
    default: <Sparkles size={20} />,
    cyberpunk: <Terminal size={20} />,
    forest: <Leaf size={20} />,
  }

  return (
    <div
      className="modal-overlay"
      onClick={handleOverlayClick}
      role="presentation"
    >
      <div
        className="settings-modal"
        ref={modalRef}
        onKeyDown={handleKeyDown}
        role="dialog"
        aria-labelledby="settings-title"
        aria-modal="true"
        tabIndex={-1}
      >
        <div className="modal-header">
          <h2 id="settings-title">⚛ Settings</h2>
          <button className="close-btn" onClick={onClose}>
            <X size={20} />
          </button>
        </div>

        <div className="modal-content">
          {/* Theme */}
          <div className="settings-section">
            <h3><Palette size={18} /> Theme</h3>
            <p className="section-description">Choose your preferred color scheme</p>
            <div className="theme-grid">
              {Object.entries(themes).map(([themeKey, theme]) => (
                <button
                  key={themeKey}
                  className={`theme-option${settings.theme === themeKey ? ' active' : ''}`}
                  onClick={() => onUpdateTheme(themeKey)}
                >
                  <div className="theme-icon">{themeIcons[themeKey] ?? <div className="theme-preview" style={{ background: theme.colors['--accent-primary'] }} />}</div>
                  <div className="theme-info">
                    <strong>{theme.name}</strong>
                    <small>{theme.description}</small>
                  </div>
                </button>
              ))}
            </div>
          </div>

          {/* Logging */}
          <div className="settings-section">
            <h3><Layout size={18} /> Logging</h3>
            <p className="section-description">Control in-memory retention, backend verbosity, and export guardrails</p>
            <div className="logging-options">
              <label htmlFor="log-retention" className="logging-label">
                <strong>Log Retention Limit</strong>
                <small>
                  Keep up to {retentionLimit.toLocaleString()} entries (default {LOG_RETENTION_DEFAULT.toLocaleString()}, soft max {LOG_RETENTION_SOFT_MAX.toLocaleString()}, hard max {HARD_MAX.toLocaleString()}).
                </small>
              </label>
              <input
                id="log-retention"
                className="logging-input"
                type="number"
                min="100"
                max={HARD_MAX}
                step="100"
                value={retentionLimit}
                onChange={(e) => onUpdateSetting('logRetentionLimit', Number(e.target.value))}
                onBlur={(e) => onUpdateSetting('logRetentionLimit', Number(e.target.value))}
              />
              {retentionWarning && (
                <p className="logging-warning">
                  Values above {LOG_RETENTION_SOFT_MAX.toLocaleString()} may affect performance; the app will enforce a hard ceiling of {HARD_MAX.toLocaleString()} entries.
                </p>
              )}
            </div>
            <div className="logging-verbosity">
              <span className="logging-subheading">Backend Verbosity</span>
              {verbosityOptions.map((opt) => (
                <div key={opt.key} className="logging-toggle">
                  <Toggle
                    checked={!!(settings.backendVerbosity?.[opt.key])}
                    onChange={(checked) => onUpdateBackendVerbosity(opt.key, checked)}
                    ariaLabel={opt.title}
                  />
                  <span className="logging-toggle-copy">
                    <strong>{opt.title}</strong>
                    <small>{opt.description}</small>
                  </span>
                </div>
              ))}
            </div>
          </div>

          {/* Profile Management */}
          <div className="settings-section">
            <h3><UserSquare size={18} /> Profile Management</h3>
            <p className="section-description">Save and load configuration presets</p>
            <div className="profile-card">
              <div className="profile-selector-wrapper">
                <div className="profile-selector">
                  <label htmlFor="profile-select">Current Profile</label>
                  <select
                    id="profile-select"
                    className="profile-dropdown"
                    value={selectedProfile}
                    onChange={(e) => onSelectProfile(e.target.value)}
                    disabled={isWatching}
                  >
                    <option value="">Select Profile</option>
                    {Object.keys(profiles || {}).map((name) => (
                      <option key={name} value={name}>{name}</option>
                    ))}
                  </select>
                </div>
                <div className="profile-buttons">
                  <button className="secondary-btn" onClick={onSaveProfile} disabled={isWatching}>
                    <Save size={16} />
                    Save Current
                  </button>
                  <button className="danger-btn" onClick={onDeleteProfile} disabled={isWatching || !selectedProfile}>
                    <Trash2 size={16} />
                    Delete Selected
                  </button>
                </div>
              </div>
            </div>
          </div>

          {/* Interface */}
          <div className="settings-section">
            <h3><Layout size={18} /> Interface</h3>
            <p className="section-description">Customize what's visible in your workspace</p>
            <div className="settings-options">
              <div className="setting-toggle">
                <Toggle
                  checked={settings.compactMode}
                  onChange={(checked) => onUpdateSetting('compactMode', checked)}
                  ariaLabel="Compact Mode"
                />
                <span className="toggle-content">
                  <strong>Compact Mode</strong>
                  <small>Reduce spacing for more content</small>
                </span>
              </div>
            </div>
          </div>

          {/* Actions */}
          <div className="settings-section">
            <div className="settings-actions">
              <button className="secondary-btn" onClick={onResetSettings}>Reset to Defaults</button>
              <button className="primary-btn" onClick={onClose}>Done</button>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
