import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, ReferenceLine } from 'recharts'

// TODO: Dev 4 — Replace with API call: getCashFlowForecast()
const mockForecast = [
  { week: 'Mar 10-16', outflow: 200000, invoices: 1, risk: 'low' },
  { week: 'Mar 17-23', outflow: 0, invoices: 0, risk: 'low' },
  { week: 'Mar 24-30', outflow: 50000, invoices: 1, risk: 'low' },
  { week: 'Mar 31-Apr 6', outflow: 0, invoices: 0, risk: 'low' },
  { week: 'Apr 7-13', outflow: 135000, invoices: 1, risk: 'medium' },
  { week: 'Apr 14-20', outflow: 75000, invoices: 1, risk: 'low' },
  { week: 'Apr 21-27', outflow: 0, invoices: 0, risk: 'low' },
  { week: 'Apr 28-May 4', outflow: 35000, invoices: 1, risk: 'low' },
  { week: 'May 5-11', outflow: 0, invoices: 0, risk: 'low' },
  { week: 'May 12-18', outflow: 0, invoices: 0, risk: 'low' },
  { week: 'May 19-25', outflow: 0, invoices: 0, risk: 'low' },
  { week: 'May 26-Jun 1', outflow: 0, invoices: 0, risk: 'low' },
]

export default function CashFlow() {
  const totalOutflows = mockForecast.reduce((sum, w) => sum + w.outflow, 0)

  return (
    <div>
      <div className="mb-8">
        <h1 className="text-2xl font-bold text-gray-900">Cash Flow Forecast</h1>
        <p className="text-gray-500 mt-1">90-day projected payment outflows</p>
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
        <div className="bg-white rounded-xl border border-gray-200 shadow-sm p-6">
          <p className="text-sm text-gray-500">Starting Balance</p>
          <p className="text-2xl font-bold text-gray-900 mt-1">₹10,00,000</p>
        </div>
        <div className="bg-white rounded-xl border border-gray-200 shadow-sm p-6">
          <p className="text-sm text-gray-500">Total Projected Outflows</p>
          <p className="text-2xl font-bold text-red-600 mt-1">-₹{totalOutflows.toLocaleString('en-IN')}</p>
        </div>
        <div className="bg-white rounded-xl border border-gray-200 shadow-sm p-6">
          <p className="text-sm text-gray-500">Projected End Balance</p>
          <p className="text-2xl font-bold text-green-600 mt-1">₹{(1000000 - totalOutflows).toLocaleString('en-IN')}</p>
        </div>
      </div>

      {/* Chart */}
      <div className="bg-white rounded-xl border border-gray-200 shadow-sm p-6">
        <h2 className="text-lg font-semibold text-gray-900 mb-6">Weekly Payment Outflows</h2>
        <div className="h-80">
          <ResponsiveContainer width="100%" height="100%">
            <BarChart data={mockForecast}>
              <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
              <XAxis
                dataKey="week"
                tick={{ fontSize: 11 }}
                angle={-35}
                textAnchor="end"
                height={80}
              />
              <YAxis
                tick={{ fontSize: 12 }}
                tickFormatter={(v) => `₹${(v / 1000).toFixed(0)}K`}
              />
              <Tooltip
                formatter={(value) => [`₹${value.toLocaleString('en-IN')}`, 'Outflow']}
                labelStyle={{ fontWeight: 'bold' }}
              />
              <Bar
                dataKey="outflow"
                fill="#3b82f6"
                radius={[4, 4, 0, 0]}
              />
            </BarChart>
          </ResponsiveContainer>
        </div>
      </div>
    </div>
  )
}
