import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Bot, ArrowLeft, RefreshCw, ExternalLink } from 'lucide-react'
import clsx from 'clsx'
import { getInvoice, getInvoiceAuditLog, reprocessInvoice } from '../services/api'

// Extract a Pine Labs payment link URL from the decision_reason string
function extractPaymentLink(text) {
  if (!text) return null
  // Try to parse JSON embedded in the string
  const jsonMatch = text.match(/\{.*\}/s)
  if (jsonMatch) {
    try {
      const parsed = JSON.parse(jsonMatch[0])
      return parsed.payment_link_url || parsed.redirect_url || parsed.checkout_url || null
    } catch {}
  }
  // Fallback: find a raw URL
  const urlMatch = text.match(/https?:\/\/[^\s"']+/)
  return urlMatch ? urlMatch[0] : null
}

const stepIcons = {
  EXTRACT: '🔍',
  VALIDATE: '✅',
  DECISION: '🤖',
  DRAFT_QUERY: '📧',
  SCHEDULE: '📅',
}

const stepLabels = {
  EXTRACT: 'Field Extraction',
  VALIDATE: 'Validation',
  DECISION: 'AI Decision',
  DRAFT_QUERY: 'Query Email',
  SCHEDULE: 'Payment Scheduling',
}

const statusStyles = {
  PENDING:    'bg-gray-100 text-gray-700',
  EXTRACTING: 'bg-blue-100 text-blue-700',
  VALIDATING: 'bg-blue-100 text-blue-700',
  APPROVED:   'bg-green-100 text-green-700',
  FLAGGED:    'bg-amber-100 text-amber-700',
  REJECTED:   'bg-red-100 text-red-700',
  SCHEDULED:  'bg-purple-100 text-purple-700',
  PAID:       'bg-emerald-100 text-emerald-700',
}

function fmt(val) {
  if (!val) return '—'
  const d = new Date(val)
  return isNaN(d.getTime()) ? val : d.toISOString().slice(0, 10)
}

export default function InvoiceDetail() {
  const { id } = useParams()
  const navigate = useNavigate()
  const [invoice, setInvoice] = useState(null)
  const [auditLog, setAuditLog] = useState([])
  const [loading, setLoading] = useState(true)
  const [reprocessing, setReprocessing] = useState(false)
  const [error, setError] = useState('')

  const load = () => {
    Promise.all([getInvoice(id), getInvoiceAuditLog(id)])
      .then(([invRes, auditRes]) => {
        if (invRes.data?.success) setInvoice(invRes.data.data)
        if (auditRes.data?.success) setAuditLog(auditRes.data.data || [])
      })
      .catch(() => setError('Failed to load invoice'))
      .finally(() => setLoading(false))
  }

  useEffect(() => {
    load()
    // Poll while invoice is still processing
    const interval = setInterval(() => {
      if (invoice && ['PENDING','EXTRACTING','VALIDATING'].includes(invoice.status)) {
        load()
      }
    }, 3000)
    return () => clearInterval(interval)
  }, [id, invoice?.status])

  const handleReprocess = async () => {
    setReprocessing(true)
    try {
      await reprocessInvoice(id)
      setTimeout(() => { load(); setReprocessing(false) }, 1500)
    } catch {
      setReprocessing(false)
    }
  }

  if (loading) return (
    <div className="flex items-center justify-center py-20">
      <p className="text-gray-500">Loading invoice...</p>
    </div>
  )

  if (error || !invoice) return (
    <div className="flex items-center justify-center py-20">
      <p className="text-red-600">{error || 'Invoice not found'}</p>
    </div>
  )

  const isProcessing = ['PENDING','EXTRACTING','VALIDATING'].includes(invoice.status)

  return (
    <div>
      {/* Header */}
      <div className="flex items-center justify-between mb-8">
        <div className="flex items-center gap-3">
          <button onClick={() => navigate('/invoices')} className="p-2 text-gray-400 hover:text-gray-600 transition-colors">
            <ArrowLeft className="w-5 h-5" />
          </button>
          <div>
            <h1 className="text-2xl font-bold text-gray-900">
              {invoice.invoice_number || 'Invoice Detail'}
            </h1>
            <p className="text-gray-500 mt-0.5 text-sm">ID: {id}</p>
          </div>
        </div>
        <div className="flex items-center gap-3">
          <span className={clsx('px-3 py-1 rounded-full text-sm font-medium', statusStyles[invoice.status] || 'bg-gray-100 text-gray-700')}>
            {invoice.status}
          </span>
          {(invoice.status === 'REJECTED' || invoice.status === 'FLAGGED') && (
            <button
              onClick={handleReprocess}
              disabled={reprocessing}
              className="flex items-center gap-2 px-4 py-2 bg-primary-600 text-white rounded-lg text-sm font-medium hover:bg-primary-700 disabled:opacity-50 transition-colors"
            >
              <RefreshCw className={clsx('w-4 h-4', reprocessing && 'animate-spin')} />
              Reprocess
            </button>
          )}
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Left — Invoice fields */}
        <div className="lg:col-span-1 space-y-6">

          {/* Pine Labs Payment Link — shown for SCHEDULED / APPROVED / PAID */}
          {['SCHEDULED', 'APPROVED', 'PAID'].includes(invoice.status) && (() => {
            const link = extractPaymentLink(invoice.decision_reason)
            return link ? (
              <div className="rounded-xl border border-green-200 bg-green-50 p-5">
                <p className="text-sm font-semibold text-green-800 mb-1">✅ Payment Scheduled</p>
                <p className="text-xs text-green-700 mb-3">
                  Pine Labs payment link generated. Share with vendor or use for early payment.
                </p>
                <a
                  href={link}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="flex items-center gap-2 w-full justify-center px-4 py-2 bg-green-600 hover:bg-green-700 text-white text-sm font-medium rounded-lg transition-colors"
                >
                  <ExternalLink className="w-4 h-4" />
                  Open Pine Labs Payment Link
                </a>
                <p className="text-xs text-green-600 mt-2 break-all">{link}</p>
              </div>
            ) : (
              <div className="rounded-xl border border-green-200 bg-green-50 p-5">
                <p className="text-sm font-semibold text-green-800 mb-1">✅ Payment Scheduled</p>
                <p className="text-xs text-green-700">
                  {invoice.decision_reason || `Payment scheduled for ${invoice.scheduled_payment_date ? new Date(invoice.scheduled_payment_date).toISOString().slice(0,10) : 'upcoming date'}`}
                </p>
              </div>
            )
          })()}

          {/* Decision reason — shown prominently for REJECTED/FLAGGED */}
          {invoice.decision_reason && (invoice.status === 'REJECTED' || invoice.status === 'FLAGGED') && (
            <div className={clsx(
              'rounded-xl border p-5',
              invoice.status === 'REJECTED' ? 'bg-red-50 border-red-200' : 'bg-amber-50 border-amber-200'
            )}>
              <p className={clsx('text-sm font-semibold mb-2', invoice.status === 'REJECTED' ? 'text-red-800' : 'text-amber-800')}>
                {invoice.status === 'REJECTED' ? '❌ Rejection Reason' : '⚠️ Flagged Reason'}
              </p>
              <p className={clsx('text-sm leading-relaxed', invoice.status === 'REJECTED' ? 'text-red-700' : 'text-amber-700')}>
                {invoice.decision_reason}
              </p>
            </div>
          )}

          {/* Invoice fields */}
          <div className="bg-white rounded-xl border border-gray-200 shadow-sm p-5">
            <h2 className="text-sm font-semibold text-gray-900 mb-4">Invoice Details</h2>
            <dl className="space-y-3">
              {[
                ['Vendor', invoice.vendor_name || invoice.vendor_id || '—'],
                ['Invoice #', invoice.invoice_number || '—'],
                ['PO Reference', invoice.po_reference || '—'],
                ['Invoice Date', fmt(invoice.invoice_date)],
                ['Due Date', fmt(invoice.due_date)],
                ['Payment Date', fmt(invoice.scheduled_payment_date)],
                ['Currency', invoice.currency || 'INR'],
                ...(invoice.pinelabs_transaction_id ? [['Pine Labs TxnID', invoice.pinelabs_transaction_id]] : []),
              ].map(([label, value]) => (
                <div key={label} className="flex justify-between text-sm">
                  <dt className="text-gray-500">{label}</dt>
                  <dd className="text-gray-900 font-medium text-right max-w-[60%] break-words">{value}</dd>
                </div>
              ))}
              <div className="pt-2 border-t border-gray-100 flex justify-between text-sm">
                <dt className="text-gray-500">Tax</dt>
                <dd className="text-gray-900 font-medium">₹{Number(invoice.tax_amount || 0).toLocaleString('en-IN')}</dd>
              </div>
              <div className="flex justify-between text-sm font-semibold">
                <dt className="text-gray-900">Total Amount</dt>
                <dd className="text-gray-900">₹{Number(invoice.total_amount || 0).toLocaleString('en-IN')}</dd>
              </div>
            </dl>
          </div>

          {/* Line items */}
          {Array.isArray(invoice.line_items) && invoice.line_items.length > 0 && (
            <div className="bg-white rounded-xl border border-gray-200 shadow-sm p-5">
              <h2 className="text-sm font-semibold text-gray-900 mb-4">Line Items</h2>
              <div className="space-y-3">
                {invoice.line_items.map((item, i) => (
                  <div key={i} className="text-sm border-b border-gray-50 pb-2 last:border-0 last:pb-0">
                    <p className="text-gray-900 font-medium">{item.description}</p>
                    <div className="flex justify-between text-gray-500 mt-0.5">
                      <span>{item.quantity} × ₹{Number(item.unit_price || 0).toLocaleString('en-IN')}</span>
                      <span className="font-medium text-gray-700">₹{Number(item.total || 0).toLocaleString('en-IN')}</span>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>

        {/* Right — Agent pipeline */}
        <div className="lg:col-span-2">
          <div className="bg-white rounded-xl border border-gray-200 shadow-sm">
            <div className="p-5 border-b border-gray-200 flex items-center gap-2">
              <Bot className="w-5 h-5 text-primary-600" />
              <h2 className="text-lg font-semibold text-gray-900">Agent Processing Pipeline</h2>
              {isProcessing && (
                <span className="ml-auto text-xs text-blue-600 bg-blue-50 px-2 py-1 rounded-full animate-pulse">
                  Processing...
                </span>
              )}
            </div>

            <div className="p-5">
              {auditLog.length === 0 ? (
                <p className="text-sm text-gray-400 text-center py-8">
                  {isProcessing ? 'Pipeline running...' : 'No audit log available'}
                </p>
              ) : (
                <div className="space-y-4">
                  {auditLog.map((step, index) => (
                    <div key={step.id || index} className="flex items-start gap-4">
                      <div className="flex flex-col items-center flex-shrink-0">
                        <div className={clsx(
                          'w-10 h-10 rounded-full flex items-center justify-center text-lg',
                          step.result === 'completed' ? 'bg-green-100' :
                          step.result === 'failed'    ? 'bg-red-100'   : 'bg-gray-100'
                        )}>
                          {stepIcons[step.step] || '⚙️'}
                        </div>
                        {index < auditLog.length - 1 && (
                          <div className={clsx('w-0.5 h-8 mt-1',
                            step.result === 'completed' ? 'bg-green-300' : 'bg-gray-200'
                          )} />
                        )}
                      </div>

                      <div className="flex-1 pb-2">
                        <div className="flex items-center justify-between gap-2">
                          <h3 className="text-sm font-semibold text-gray-900">
                            {stepLabels[step.step] || step.step}
                          </h3>
                          <div className="flex items-center gap-2 flex-shrink-0">
                            {step.duration_ms > 0 && (
                              <span className="text-xs text-gray-400">{step.duration_ms}ms</span>
                            )}
                            <span className={clsx(
                              'px-2 py-0.5 rounded-full text-xs font-medium',
                              step.result === 'completed' ? 'bg-green-100 text-green-700' :
                              step.result === 'failed'    ? 'bg-red-100 text-red-700'     :
                              'bg-yellow-100 text-yellow-700'
                            )}>
                              {step.result}
                            </span>
                          </div>
                        </div>

                        <p className="text-sm text-gray-600 mt-1 leading-relaxed">{step.reasoning}</p>

                        {step.confidence_score > 0 && (
                          <div className="mt-2 flex items-center gap-2">
                            <div className="w-24 h-1.5 bg-gray-200 rounded-full overflow-hidden">
                              <div
                                className={clsx('h-full rounded-full',
                                  step.confidence_score >= 0.8 ? 'bg-green-500' :
                                  step.confidence_score >= 0.5 ? 'bg-amber-500' : 'bg-red-500'
                                )}
                                style={{ width: `${step.confidence_score * 100}%` }}
                              />
                            </div>
                            <span className="text-xs text-gray-400">
                              {(step.confidence_score * 100).toFixed(0)}% confidence
                            </span>
                          </div>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
