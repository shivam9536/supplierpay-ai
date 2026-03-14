import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { FileText, Eye } from 'lucide-react'
import clsx from 'clsx'
import { getInvoices } from '../services/api'

const statusStyles = {
  PENDING: 'bg-gray-100 text-gray-700',
  EXTRACTING: 'bg-blue-100 text-blue-700',
  VALIDATING: 'bg-blue-100 text-blue-700',
  APPROVED: 'bg-green-100 text-green-700',
  FLAGGED: 'bg-amber-100 text-amber-700',
  REJECTED: 'bg-red-100 text-red-700',
  SCHEDULED: 'bg-purple-100 text-purple-700',
  PAID: 'bg-emerald-100 text-emerald-700',
}

function formatDate(val) {
  if (!val) return '—'
  const d = typeof val === 'string' ? new Date(val) : val
  return isNaN(d.getTime()) ? '—' : d.toISOString().slice(0, 10)
}

export default function Invoices() {
  const navigate = useNavigate()
  const [invoices, setInvoices] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const fetchInvoices = () => {
    getInvoices()
      .then((res) => {
        if (res.data?.success && Array.isArray(res.data.data)) setInvoices(res.data.data)
      })
      .catch(() => setError('Failed to load invoices'))
      .finally(() => setLoading(false))
  }

  useEffect(() => {
    fetchInvoices()
    const interval = setInterval(fetchInvoices, 5000)
    return () => clearInterval(interval)
  }, [])

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <p className="text-gray-500">Loading invoices...</p>
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex items-center justify-center py-12">
        <p className="text-red-600">{error}</p>
      </div>
    )
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Invoices</h1>
          <p className="text-gray-500 mt-1">All processed invoices and their agent decisions</p>
        </div>
        <button
          onClick={() => navigate('/upload')}
          className="px-4 py-2 bg-primary-600 text-white rounded-lg text-sm font-medium hover:bg-primary-700 transition-colors"
        >
          + Upload Invoice
        </button>
      </div>

      {/* Invoice Table */}
      {invoices.length === 0 ? (
        <div className="bg-white rounded-xl border border-gray-200 shadow-sm p-16 text-center">
          <FileText className="w-12 h-12 text-gray-300 mx-auto mb-4" />
          <p className="text-gray-500 font-medium">No invoices yet</p>
          <p className="text-sm text-gray-400 mt-1">Upload a PDF invoice to get started</p>
          <button
            onClick={() => navigate('/upload')}
            className="mt-4 px-4 py-2 bg-primary-600 text-white rounded-lg text-sm font-medium hover:bg-primary-700 transition-colors"
          >
            Upload Invoice
          </button>
        </div>
      ) : (
      <div className="bg-white rounded-xl border border-gray-200 shadow-sm overflow-hidden">
        <table className="w-full">
          <thead className="bg-gray-50 border-b border-gray-200">
            <tr>
              <th className="text-left px-6 py-3 text-xs font-medium text-gray-500 uppercase">Invoice #</th>
              <th className="text-left px-6 py-3 text-xs font-medium text-gray-500 uppercase">Vendor</th>
              <th className="text-right px-6 py-3 text-xs font-medium text-gray-500 uppercase">Amount</th>
              <th className="text-center px-6 py-3 text-xs font-medium text-gray-500 uppercase">Status</th>
              <th className="text-left px-6 py-3 text-xs font-medium text-gray-500 uppercase">Due Date</th>
              <th className="text-left px-6 py-3 text-xs font-medium text-gray-500 uppercase">Payment Date</th>
              <th className="text-center px-6 py-3 text-xs font-medium text-gray-500 uppercase">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-100">
            {invoices.map((inv) => (
              <tr key={inv.id} className="hover:bg-gray-50 transition-colors">
                <td className="px-6 py-4">
                  <div className="flex items-center gap-2">
                    <FileText className="w-4 h-4 text-gray-400" />
                    <span className="text-sm font-medium text-gray-900">{inv.invoice_number || '—'}</span>
                  </div>
                </td>
                <td className="px-6 py-4 text-sm text-gray-600">{inv.vendor_name || inv.vendor_id || '—'}</td>
                <td className="px-6 py-4 text-sm text-gray-900 text-right font-medium">
                  ₹{Number(inv.total_amount || 0).toLocaleString('en-IN')}
                </td>
                <td className="px-6 py-4 text-center">
                  <span className={clsx('px-2.5 py-1 rounded-full text-xs font-medium', statusStyles[inv.status] || 'bg-gray-100 text-gray-700')}>
                    {inv.status}
                  </span>
                </td>
                <td className="px-6 py-4 text-sm text-gray-600">{formatDate(inv.due_date)}</td>
                <td className="px-6 py-4 text-sm text-gray-600">{formatDate(inv.scheduled_payment_date)}</td>
                <td className="px-6 py-4 text-center">
                  <button
                    onClick={() => navigate(`/invoices/${inv.id}`)}
                    className="p-1.5 text-gray-400 hover:text-primary-600 transition-colors"
                    title="View details"
                  >
                    <Eye className="w-4 h-4" />
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      )}
    </div>
  )
}
