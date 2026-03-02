import React, { useState } from 'react'
import AdvancedOperationsPanel from './AdvancedOperationsPanel'
import type { ErrorPayload } from '../../lib/errorBus'
import type { Config } from '../../context/AppContext'

interface Props {
  config: Config
  onError: (payload: ErrorPayload) => void
}

export default function AdvancedOperationsView({ config, onError }: Props) {
  const [panelError, setPanelError] = useState<ErrorPayload | null>(null)

  const handlePanelError = (payload: ErrorPayload) => {
    setPanelError(payload)
    onError(payload)
  }

  return (
    <div className="operations-view">
      <div className="panel-wrapper">
        {panelError && (
          <div className="error-banner">
            <div>
              <strong>{panelError.context || 'Advanced operations'}</strong>
              <span>{panelError.message || 'An unexpected error occurred.'}</span>
            </div>
            <button type="button" onClick={() => setPanelError(null)}>Dismiss</button>
          </div>
        )}
        <AdvancedOperationsPanel config={config} onError={handlePanelError} />
      </div>
    </div>
  )
}
