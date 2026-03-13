import { useParams } from 'react-router-dom'
import { CheckCircle, XCircle, Clock, AlertTriangle, Bot } from 'lucide-react'
import clsx from 'clsx'

// TODO: Dev 4 — Replace with API calls: getInvoice(id), getInvoiceAuditLog(id)
// Also implement SSE subscription for live updates

const stepIcons = {
  EXTRACT: '🔍',
  VALIDATE: '✅',
  CROSS_REFERENCE: '🔗',
  DECISION: '🤖',
  DRAFT_QUERY: '📧',
  SCHEDULE: '📅',
  DISBURSE: '💰',
}

const stepLabels = {
  EXTRACT: 'Field Extraction',
  VALIDATE: 'Validation',
  CROSS_REFERENCE: 'PO Cross-Reference',
  DECISION: 'AI Decision',
  DRAFT_QUERY: 'Query Email',
  SCHEDULE: 'Payment Scheduling',
  DISBURSE: 'Disbursement',
}

export default function InvoiceDetail() {
  const { id } = useParams()

  // Mock audit log — will be replaced with real data
  const auditSteps = [
    { step: 'EXTRACT', result: 'completed', reasoning: 'Extracted 7 fields with 95% average confidence', confidence: 0.95, duration: 1250 },
    { step: 'VALIDATE', result: 'completed', reasoning: 'All required fields present, amounts valid, dates in range', confidence: 1.0, duration: 45 },
    { step: 'CROSS_REFERENCE', result: 'completed', reasoning: 'PO-2026-100 matched. Amount: ₹50,000 = PO value. Line items: 2/2 matched.', confidence: 1.0, duration: 120 },
    { step: 'DECISION', result: 'completed', reasoning: 'Auto-approved: All checks passed.', confidence: 1.0, duration: 30 },
    { step: 'SCHEDULE', result: 'completed', reasoning: 'Payment scheduled for Day 28 (2026-03-29). No early discount benefit.', confidence: 1.0, duration: 15 },
  ]

  return (
    <div>
      <div className="mb-8">
        <h1 className="text-2xl font-bold text-gray-900">Invoice Detail</h1>
        <p className="text-gray-500 mt-1">ID: {id}</p>
      </div>

      {/* Agent Pipeline Visualization */}
      <div className="bg-white rounded-xl border border-gray-200 shadow-sm mb-8">
        <div className="p-6 border-b border-gray-200 flex items-center gap-2">
          <Bot className="w-5 h-5 text-primary-600" />
          <h2 className="text-lg font-semibold text-gray-900">Agent Processing Pipeline</h2>
        </div>
        <div className="p-6">
          <div className="space-y-4">
            {auditSteps.map((step, index) => (
              <div key={step.step} className="flex items-start gap-4">
                {/* Step indicator */}
                <div className="flex flex-col items-center">
                  <div className={clsx(
                    'w-10 h-10 rounded-full flex items-center justify-center text-lg',
                    step.result === 'completed' ? 'bg-green-100' : 
                    step.result === 'failed' ? 'bg-red-100' : 'bg-gray-100'
                  )}>
                    {stepIcons[step.step]}
                  </div>
                  {index < auditSteps.length - 1 && (
                    <div className="w-0.5 h-8 bg-green-300 mt-1" />
                  )}
                </div>

                {/* Step content */}
                <div className="flex-1 pb-4">
                  <div className="flex items-center justify-between">
                    <h3 className="text-sm font-semibold text-gray-900">
                      {stepLabels[step.step]}
                    </h3>
                    <div className="flex items-center gap-3">
                      <span className="text-xs text-gray-400">{step.duration}ms</span>
                      <span className={clsx(
                        'px-2 py-0.5 rounded-full text-xs font-medium',
                        step.result === 'completed' ? 'bg-green-100 text-green-700' :
                        step.result === 'failed' ? 'bg-red-100 text-red-700' :
                        'bg-yellow-100 text-yellow-700'
                      )}>
                        {step.result}
                      </span>
                    </div>
                  </div>
                  <p className="text-sm text-gray-600 mt-1">{step.reasoning}</p>
                  {step.confidence && (
                    <div className="mt-2 flex items-center gap-2">
                      <div className="w-24 h-1.5 bg-gray-200 rounded-full overflow-hidden">
                        <div
                          className="h-full bg-green-500 rounded-full"
                          style={{ width: `${step.confidence * 100}%` }}
                        />
                      </div>
                      <span className="text-xs text-gray-400">
                        {(step.confidence * 100).toFixed(0)}% confidence
                      </span>
                    </div>
                  )}
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}
