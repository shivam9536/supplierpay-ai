import { FileText, CheckCircle, AlertTriangle, Clock, CreditCard, TrendingUp } from 'lucide-react'

const stats = [
  { label: 'Total Invoices', value: '3', icon: FileText, color: 'text-blue-600', bg: 'bg-blue-50' },
  { label: 'Approved', value: '1', icon: CheckCircle, color: 'text-green-600', bg: 'bg-green-50' },
  { label: 'Flagged', value: '1', icon: AlertTriangle, color: 'text-amber-600', bg: 'bg-amber-50' },
  { label: 'Scheduled', value: '1', icon: Clock, color: 'text-purple-600', bg: 'bg-purple-50' },
  { label: 'Payments Due', value: '₹2.5L', icon: CreditCard, color: 'text-red-600', bg: 'bg-red-50' },
  { label: 'Saved (Discounts)', value: '₹5,000', icon: TrendingUp, color: 'text-emerald-600', bg: 'bg-emerald-50' },
]

export default function Dashboard() {
  return (
    <div>
      <div className="mb-8">
        <h1 className="text-2xl font-bold text-gray-900">Dashboard</h1>
        <p className="text-gray-500 mt-1">Overview of your accounts payable pipeline</p>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mb-8">
        {stats.map(({ label, value, icon: Icon, color, bg }) => (
          <div key={label} className="bg-white rounded-xl p-6 border border-gray-200 shadow-sm">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-500">{label}</p>
                <p className="text-2xl font-bold text-gray-900 mt-1">{value}</p>
              </div>
              <div className={`p-3 rounded-lg ${bg}`}>
                <Icon className={`w-6 h-6 ${color}`} />
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* Recent Activity */}
      <div className="bg-white rounded-xl border border-gray-200 shadow-sm">
        <div className="p-6 border-b border-gray-200">
          <h2 className="text-lg font-semibold text-gray-900">Recent Agent Activity</h2>
        </div>
        <div className="p-6">
          {/* TODO: Dev 4 — Implement real-time activity feed */}
          <div className="space-y-4">
            <ActivityItem
              step="APPROVED"
              message="Invoice INV-ACME-2026-042 auto-approved. Payment scheduled for Mar 29."
              time="2 hours ago"
              type="success"
            />
            <ActivityItem
              step="FLAGGED"
              message="Invoice INV-TP-2026-118 flagged: Amount exceeds PO by 8%. Query email sent."
              time="3 hours ago"
              type="warning"
            />
            <ActivityItem
              step="SCHEDULED"
              message="Invoice INV-SN-2026-007 scheduled for early payment (Mar 15). Saving ₹5,000 discount."
              time="5 hours ago"
              type="info"
            />
          </div>
        </div>
      </div>
    </div>
  )
}

function ActivityItem({ step, message, time, type }) {
  const colors = {
    success: 'border-green-400 bg-green-50',
    warning: 'border-amber-400 bg-amber-50',
    info: 'border-blue-400 bg-blue-50',
  }

  return (
    <div className={`border-l-4 p-4 rounded-r-lg ${colors[type]}`}>
      <p className="text-sm text-gray-800">{message}</p>
      <p className="text-xs text-gray-500 mt-1">{time}</p>
    </div>
  )
}
