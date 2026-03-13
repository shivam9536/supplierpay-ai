export default function PaymentSchedule() {
  // TODO: Dev 4 — Replace with API call: getPaymentSchedule()
  const mockSchedule = [
    { date: '2026-03-15', vendor: 'SecureNet Cybersecurity', invoice: 'INV-SN-2026-007', amount: 200000, reason: 'Early payment — 2.5% discount (₹5,000 saved)' },
    { date: '2026-03-29', vendor: 'Acme Cloud Solutions', invoice: 'INV-ACME-2026-042', amount: 50000, reason: 'Day 28 of net-30 — maximise cash float' },
  ]

  return (
    <div>
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Payment Schedule</h1>
          <p className="text-gray-500 mt-1">Upcoming automated payments optimised for cash flow</p>
        </div>
        <button className="px-4 py-2 bg-primary-600 text-white rounded-lg text-sm font-medium hover:bg-primary-700 transition-colors">
          Trigger Payment Run
        </button>
      </div>

      <div className="space-y-4">
        {mockSchedule.map((payment) => (
          <div key={payment.invoice} className="bg-white rounded-xl border border-gray-200 shadow-sm p-6">
            <div className="flex items-center justify-between">
              <div>
                <div className="flex items-center gap-3">
                  <span className="text-lg font-bold text-gray-900">
                    ₹{payment.amount.toLocaleString('en-IN')}
                  </span>
                  <span className="px-2.5 py-1 rounded-full text-xs font-medium bg-purple-100 text-purple-700">
                    SCHEDULED
                  </span>
                </div>
                <p className="text-sm text-gray-600 mt-1">{payment.vendor} — {payment.invoice}</p>
                <p className="text-sm text-gray-500 mt-1">📅 Payment date: {payment.date}</p>
              </div>
              <div className="text-right">
                <p className="text-sm text-gray-500 bg-gray-50 rounded-lg p-3 max-w-xs">
                  💡 {payment.reason}
                </p>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
