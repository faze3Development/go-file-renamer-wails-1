import React, { useState } from 'react'
import { Folder, FolderOpen, Eye, FileText, Cog, Play, Square, Info } from 'lucide-react'
import type { Config, NamerInfo, ActionInfo, PatternInfo } from '../context/AppContext'

interface Props {
  config: Config
  watchPathDisplay: string
  selectedPatternID: string
  availablePatterns: PatternInfo[]
  availableNamers: NamerInfo[]
  availableActions: ActionInfo[]
  isWatching: boolean
  isBusy: boolean
  onSelectDirectory: () => void
  onPatternSelect: (id: string) => void
  onUpdateConfig: (key: keyof Config, value: unknown) => void
  onSelectActionDirectory: () => void
  onRequestStart: () => void
  onRequestStop: () => void
  onOpenAdvancedOperations: () => void
  onOpenMonitoring: () => void
}

const TEMPLATE_PLACEHOLDER = '{original}-{date}'

export default function ConfigurationView({
  config, watchPathDisplay, selectedPatternID, availablePatterns, availableNamers, availableActions,
  isWatching, isBusy, onSelectDirectory, onPatternSelect, onUpdateConfig, onSelectActionDirectory,
  onRequestStart, onRequestStop, onOpenAdvancedOperations, onOpenMonitoring,
}: Props) {
  const [showPatternInfo, setShowPatternInfo] = useState(false)
  const [showNamingInfo, setShowNamingInfo] = useState(false)
  const [showActionInfo, setShowActionInfo] = useState(false)

  const statusTone = isWatching ? 'active' : isBusy ? 'busy' : 'idle'
  const statusMessage = isWatching
    ? 'Watching for file changes...'
    : isBusy
    ? 'Processing pending actions...'
    : 'Watcher idle'

  return (
    <div className="config-section">
      {/* Directory Selection */}
      <div className="config-card directory-selection">
        <div className="card-header">
          <h3><Folder size={16} />Watch Directory</h3>
          <p>Select or drag &amp; drop a folder to monitor</p>
        </div>
        <div className="card-content">
          <div className="directory-input">
            <input
              id="watch-path"
              type="text"
              readOnly
              value={watchPathDisplay}
              placeholder="No directory selected..."
              className="directory-path"
            />
            <button className="browse-btn" onClick={onSelectDirectory}>
              <FolderOpen size={16} />
              Browse
            </button>
            {!isWatching ? (
              <button className="start-btn" onClick={onRequestStart} disabled={isBusy}>
                <Play size={16} />
                Start Watching
              </button>
            ) : (
              <button className="stop-btn" onClick={onRequestStop} disabled={isBusy}>
                <Square size={16} />
                Stop Watching
              </button>
            )}
          </div>
          <div className="status-row">
            <span className={`status-indicator ${statusTone}`}>{statusMessage}</span>
            <button className="status-link" type="button" onClick={onOpenMonitoring}>
              View monitoring
            </button>
          </div>
        </div>
      </div>

      {/* Pattern & Naming Grid */}
      <div className="pattern-naming-grid">
        {/* Pattern Matching */}
        <div className="config-card compact">
          <div className="card-header">
            <h3><Eye size={16} />Pattern Matching</h3>
            <button
              className="info-btn"
              onMouseEnter={() => setShowPatternInfo(true)}
              onMouseLeave={() => setShowPatternInfo(false)}
              aria-label="Show additional information about pattern matching"
            >
              <Info size={14} />
            </button>
          </div>
          {showPatternInfo && (
            <div className="info-tooltip">Define which files to rename using pattern matching</div>
          )}
          <div className="card-content">
            <div className="form-group compact">
              <select
                id="pattern-id"
                value={selectedPatternID}
                onChange={(e) => onPatternSelect(e.target.value)}
              >
                {availablePatterns.length === 0 ? (
                  <option value="">Loading patterns...</option>
                ) : (
                  <>
                    {availablePatterns.map((p) => (
                      <option key={p.id} value={p.id} title={p.description}>{p.name}</option>
                    ))}
                    <option value="custom">Custom Regex</option>
                  </>
                )}
              </select>
            </div>
            {selectedPatternID === 'custom' && (
              <div className="form-group compact">
                <input
                  id="name-pattern"
                  type="text"
                  value={config.NamePattern}
                  onChange={(e) => onUpdateConfig('NamePattern', e.target.value)}
                  placeholder="Custom regex pattern"
                />
              </div>
            )}
          </div>
        </div>

        {/* Naming Scheme */}
        <div className="config-card compact">
          <div className="card-header">
            <h3><FileText size={16} />Naming Scheme</h3>
            <button
              className="info-btn"
              onMouseEnter={() => setShowNamingInfo(true)}
              onMouseLeave={() => setShowNamingInfo(false)}
              aria-label="Show additional information about naming schemes"
            >
              <Info size={14} />
            </button>
          </div>
          {showNamingInfo && (
            <div className="info-tooltip">Choose how to rename matched files</div>
          )}
          <div className="card-content">
            <div className="form-group compact">
              <select
                id="namer-id"
                value={config.NamerID}
                onChange={(e) => onUpdateConfig('NamerID', e.target.value)}
              >
                {availableNamers.length === 0 ? (
                  <option value="">Loading naming methods...</option>
                ) : (
                  availableNamers.map((n) => (
                    <option key={n.id} value={n.id} title={n.description}>{n.name}</option>
                  ))
                )}
              </select>
            </div>
            {config.NamerID === 'random' && (
              <div className="form-group compact">
                <input
                  id="name-length"
                  type="number"
                  value={config.RandomLength}
                  onChange={(e) => onUpdateConfig('RandomLength', Number(e.target.value))}
                  placeholder="Random name length"
                />
              </div>
            )}
            {config.NamerID === 'template' && (
              <div className="form-group compact">
                <input
                  id="template-string"
                  type="text"
                  value={config.TemplateString}
                  onChange={(e) => onUpdateConfig('TemplateString', e.target.value)}
                  placeholder={`e.g. ${TEMPLATE_PLACEHOLDER}`}
                />
                <div className="help-text compact">
                  {'{'}original{'}'}, {'{'}date{'}'}, {'{'}time{'}'}, {'{'}count{'}'}
                </div>
              </div>
            )}
            {(config.NamerID === 'custom_datetime' || config.NamerID === 'sequential_datetime') && (
              <div className="form-group compact">
                <input
                  id="datetime-format"
                  type="text"
                  value={config.DateTimeFormat}
                  onChange={(e) => onUpdateConfig('DateTimeFormat', e.target.value)}
                  placeholder="2006-01-02_15-04-05"
                />
                <div className="help-text compact">
                  Go format: 2006=year, 01=month, 02=day
                </div>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Post-Rename Actions */}
      <div className="config-card compact full-width">
        <div className="card-header">
          <h3><Cog size={16} />Post-Rename Actions</h3>
          <button
            className="info-btn"
            onMouseEnter={() => setShowActionInfo(true)}
            onMouseLeave={() => setShowActionInfo(false)}
            aria-label="Show additional information about post-rename actions"
          >
            <Info size={14} />
          </button>
        </div>
        {showActionInfo && (
          <div className="info-tooltip">What to do after renaming files</div>
        )}
        <div className="card-content">
          <div className="form-group compact">
            <select
              id="action-id"
              value={config.ActionID}
              onChange={(e) => onUpdateConfig('ActionID', e.target.value)}
            >
              {availableActions.length === 0 ? (
                <option value="">Loading actions...</option>
              ) : (
                availableActions.map((a) => (
                  <option key={a.id} value={a.id} title={a.description}>{a.name}</option>
                ))
              )}
            </select>
          </div>
          {(config.ActionID === 'move' || config.ActionID === 'copy') && (
            <div className="form-group compact">
              <div className="directory-input">
                <input
                  id="action-dest"
                  type="text"
                  readOnly
                  placeholder="Select destination..."
                  value={config.ActionConfig?.destinationPath || ''}
                  className="directory-path"
                />
                <button className="browse-btn" onClick={onSelectActionDirectory}>
                  <FolderOpen size={16} />
                  Browse
                </button>
              </div>
            </div>
          )}
          {config.ActionID === 'advanced_operations' && (
            <div className="action-info-card compact">
              <button className="secondary-btn" onClick={onOpenAdvancedOperations} disabled={isWatching}>
                Manage Advanced Operations
              </button>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
