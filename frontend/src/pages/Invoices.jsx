import { useNavigate } from 'react-router-dom'
import { FileText, Eye } from 'lucide-react'
import clsx from 'clsx'

// Mock data — will be replaced with API calls
const mockInvoices = [
  {
    id: 'aaaa1111-1111-1111-1111-111111111111',
    invoice_number: 'INV-ACME-2026-042',
    vendor_name: 'Acme Cloud Solutions',
    total_amount: 50000,
    status: 'APPROVED',
    due_date: '2026-03-31',
    scheduled_payment_date: '2026-03-29',
  },
  {
    id: 'bbbb2222-2222-2222-2222-222222222222',
    invoice_number: 'INV-TP-2026-118',
    vendor_name: 'TechParts India Pvt Ltd',
    total_amount: 135000,
    status: 'FLAGGED',
    due_date: '2026-04-07',
    scheduled_payment_date: null,
  },
  {
    id: 'cccc3333-3333-3333-3333-333333333333',
    invoice_number: 'INV-SN-2026-007',
    vendor_name: 'SecureNet Cybersecurity',
    total_amount: 200000,
    status: 'SCHEDULED',
    due_date: '2026-04-04',
    scheduled_payment_date: '2026-03-15',
  },
]

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

export default function Invoices() {
  const navigate = useNavigate()

  // TODO: Dev 4 — Replace with API call: getInvoices()

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
            {mockInvoices.map((inv) => (
              <tr key={inv.id} className="hover:bg-gray-50 transition-colors">
                <td className="px-6 py-4">
                  <div className="flex items-center gap-2">
                    <FileText className="w-4 h-4 text-gray-400" />
                    <span className="text-sm font-medium text-gray-900">{inv.invoice_number}</span>
                  </div>
                </td>
                <td className="px-6 py-4 text-sm text-gray-600">{inv.vendor_name}</td>
                <td className="px-6 py-4 text-sm text-gray-900 text-right font-medium">
                  ₹{inv.total_amount.toLocaleString('en-IN')}
                </td>
                <td className="px-6 py-4 text-center">
                  <span className={clsx('px-2.5 py-1 rounded-full text-xs font-medium', statusStyles[inv.status])}>
                    {inv.status}
                  </span>
                </td>
                <td className="px-6 py-4 text-sm text-gray-600">{inv.due_date}</td>
                <td className="px-6 py-4 text-sm text-gray-600">{inv.scheduled_payment_date || '—'}</td>
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
    </div>
  )
}
