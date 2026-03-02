import React, { useMemo } from 'react'
import Toggle from '../shared/Toggle'
import type { Config } from '../../context/AppContext'

interface Props {
  config: Config
  isWatching: boolean
  onUpdateConfig: (key: keyof Config, value: unknown) => void
}

export default function AdvancedView({ config, isWatching, onUpdateConfig }: Props) {
  const preview = useMemo(() => {
    try {
      const now = new Date()
      const year = now.getFullYear().toString()
      const month = String(now.getMonth() + 1).padStart(2, '0')
      const day = String(now.getDate()).padStart(2, '0')
      const hour = String(now.getHours()).padStart(2, '0')
      const minute = String(now.getMinutes()).padStart(2, '0')
      const second = String(now.getSeconds()).padStart(2, '0')
      const date = now.toISOString().slice(0, 10)
      const time = now.toTimeString().slice(0, 5).replace(':', '-')
      const datetime = now.toISOString().slice(0, 19).replace(/:/g, '-').replace('T', '_')
      const unix = Math.floor(Date.now() / 1000).toString()
      const unixmilli = Date.now().toString()

      if (config.NamerID === 'datetime') return datetime
      if (config.NamerID === 'custom_datetime') {
        return config.DateTimeFormat
          .replace('2006', year).replace('01', month).replace('02', day)
          .replace('15', hour).replace('04', minute).replace('05', second)
      }
      if (config.NamerID === 'template') {
        return config.TemplateString
          .replace('{original}', 'example').replace('{date}', date).replace('{time}', time)
          .replace('{datetime}', datetime).replace('{year}', year).replace('{month}', month)
          .replace('{day}', day).replace('{hour}', hour).replace('{minute}', minute)
          .replace('{second}', second).replace('{unix}', unix).replace('{unixmilli}', unixmilli)
          .replace('{count}', '1').replace('{count:2}', '01').replace('{count:3}', '001').replace('{count:4}', '0001')
      }
      return 'example.jpg'
    } catch {
      return 'Error generating preview'
    }
  }, [config.NamerID, config.DateTimeFormat, config.TemplateString])

  return (
    <div className="advanced-section">
      {/* Directory Processing */}
      <div className="config-card">
        <div className="card-header">
          <h3>📁 Directory Processing</h3>
          <p>Configure how files and directories are processed</p>
        </div>
        <div className="card-content">
          <div className="option-toggle">
            <Toggle
              checked={config.Recursive}
              disabled={isWatching}
              onChange={(checked) => onUpdateConfig('Recursive', checked)}
              ariaLabel="Watch Recursively"
            />
            <span className="toggle-label">
              <strong>Watch Recursively</strong>
              <small>Monitor subdirectories and their contents</small>
            </span>
          </div>
        </div>
      </div>

      {/* Execution Mode */}
      <div className="config-card">
        <div className="card-header">
          <h3>⚙️ Execution Mode</h3>
          <p>Control how file operations are performed</p>
        </div>
        <div className="card-content">
          <div className="options-grid">
            <div className="option-toggle">
              <Toggle
                checked={config.DryRun}
                disabled={isWatching}
                onChange={(checked) => onUpdateConfig('DryRun', checked)}
                ariaLabel="Dry Run Mode"
              />
              <span className="toggle-label">
                <strong>Dry Run Mode</strong>
                <small>Preview changes without actually renaming files</small>
              </span>
            </div>
            <div className="option-toggle">
              <Toggle
                checked={config.NoInitialScan}
                disabled={isWatching}
                onChange={(checked) => onUpdateConfig('NoInitialScan', checked)}
                ariaLabel="Skip Initial Scan"
              />
              <span className="toggle-label">
                <strong>Skip Initial Scan</strong>
                <small>Only process new files, ignore existing ones</small>
              </span>
            </div>
          </div>
        </div>
      </div>

      {/* Date/Time Preview */}
      <div className="config-card">
        <div className="card-header">
          <h3>📅 Date &amp; Time Preview</h3>
          <p>Preview how your date/time naming will look</p>
        </div>
        <div className="card-content">
          <div className="datetime-preview">
            <div className="preview-item">
              <label htmlFor="filename-preview">Sample Filename:</label>
              <div className="filename-preview" id="filename-preview">{preview}</div>
            </div>
          </div>
        </div>
      </div>

      {/* Performance Settings */}
      <div className="config-card">
        <div className="card-header">
          <h3>🚀 Performance Settings</h3>
          <p>Fine-tune processing behavior and timing</p>
        </div>
        <div className="card-content">
          <div className="advanced-form-grid">
            <div className="form-group">
              <label htmlFor="settle-time">File Settle Time (ms)</label>
              <input
                id="settle-time"
                type="number"
                value={config.Settle}
                onChange={(e) => onUpdateConfig('Settle', Number(e.target.value))}
                disabled={isWatching}
                min="0"
                max="5000"
                step="50"
              />
              <div className="help-text">Time to wait for file operations to complete</div>
            </div>
            <div className="form-group">
              <label htmlFor="settle-timeout">Settle Timeout (seconds)</label>
              <input
                id="settle-timeout"
                type="number"
                value={config.SettleTimeout}
                onChange={(e) => onUpdateConfig('SettleTimeout', Number(e.target.value))}
                disabled={isWatching}
                min="1"
                max="30"
              />
              <div className="help-text">Maximum wait time for file stability</div>
            </div>
            <div className="form-group">
              <label htmlFor="retry-count">Retry Attempts</label>
              <input
                id="retry-count"
                type="number"
                value={config.Retries}
                onChange={(e) => onUpdateConfig('Retries', Number(e.target.value))}
                disabled={isWatching}
                min="1"
                max="10"
              />
              <div className="help-text">Number of retry attempts for failed operations</div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
