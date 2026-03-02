import React from 'react'

interface Step {
  label: string
  hint: string
  done: boolean
}

interface Props {
  steps: Step[]
}

export default function SetupStepper({ steps }: Props) {
  const firstIncomplete = steps.findIndex(s => !s.done)
  const activeIndex = firstIncomplete === -1 ? steps.length - 1 : firstIncomplete

  return (
    <div className="setup-stepper">
      {steps.map((step, i) => {
        const isDone   = step.done
        const isActive = i === activeIndex && !isDone
        return (
          <React.Fragment key={i}>
            <div className={`stepper-step${isDone ? ' done' : isActive ? ' active' : ''}`}>
              <div className="stepper-circle">
                {isDone ? (
                  <svg width="12" height="12" viewBox="0 0 12 12" fill="none">
                    <path d="M2 6l3 3 5-5" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                  </svg>
                ) : (
                  <span>{i + 1}</span>
                )}
              </div>
              <div className="stepper-info">
                <span className="stepper-label">{step.label}</span>
                <span className="stepper-hint">{step.hint}</span>
              </div>
            </div>
            {i < steps.length - 1 && (
              <div className={`stepper-line${step.done ? ' done' : ''}`} />
            )}
          </React.Fragment>
        )
      })}
    </div>
  )
}
