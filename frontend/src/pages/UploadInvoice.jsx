import { useState } from 'react'
import { Upload, FileText, Loader2 } from 'lucide-react'

export default function UploadInvoice() {
  const [file, setFile] = useState(null)
  const [uploading, setUploading] = useState(false)
  const [result, setResult] = useState(null)
  const [dragOver, setDragOver] = useState(false)

  const handleDrop = (e) => {
    e.preventDefault()
    setDragOver(false)
    const dropped = e.dataTransfer.files[0]
    if (dropped) setFile(dropped)
  }

  const handleUpload = async () => {
    if (!file) return
    setUploading(true)

    // TODO: Dev 4 — Call uploadInvoice() API with FormData
    // Then subscribe to SSE for real-time pipeline updates
    setTimeout(() => {
      setResult({
        invoice_id: 'new-invoice-id',
        status: 'PENDING',
        message: 'Invoice uploaded and processing started',
      })
      setUploading(false)
    }, 1500)
  }

  return (
    <div>
      <div className="mb-8">
        <h1 className="text-2xl font-bold text-gray-900">Upload Invoice</h1>
        <p className="text-gray-500 mt-1">Upload a PDF invoice or paste JSON payload</p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
        {/* PDF Upload */}
        <div className="bg-white rounded-xl border border-gray-200 shadow-sm p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">PDF Upload</h2>
          <div
            onDragOver={(e) => { e.preventDefault(); setDragOver(true) }}
            onDragLeave={() => setDragOver(false)}
            onDrop={handleDrop}
            className={`border-2 border-dashed rounded-xl p-12 text-center transition-colors ${
              dragOver ? 'border-primary-400 bg-primary-50' : 'border-gray-300 hover:border-gray-400'
            }`}
          >
            <Upload className="w-12 h-12 text-gray-400 mx-auto mb-4" />
            <p className="text-gray-600 mb-2">Drag & drop your invoice PDF here</p>
            <p className="text-sm text-gray-400 mb-4">or</p>
            <label className="px-4 py-2 bg-primary-600 text-white rounded-lg text-sm font-medium cursor-pointer hover:bg-primary-700 transition-colors">
              Browse Files
              <input
                type="file"
                accept=".pdf"
                className="hidden"
                onChange={(e) => setFile(e.target.files[0])}
              />
            </label>
          </div>

          {file && (
            <div className="mt-4 p-4 bg-gray-50 rounded-lg flex items-center justify-between">
              <div className="flex items-center gap-3">
                <FileText className="w-5 h-5 text-primary-600" />
                <div>
                  <p className="text-sm font-medium text-gray-900">{file.name}</p>
                  <p className="text-xs text-gray-500">{(file.size / 1024).toFixed(1)} KB</p>
                </div>
              </div>
              <button
                onClick={handleUpload}
                disabled={uploading}
                className="px-4 py-2 bg-primary-600 text-white rounded-lg text-sm font-medium hover:bg-primary-700 disabled:opacity-50 flex items-center gap-2 transition-colors"
              >
                {uploading && <Loader2 className="w-4 h-4 animate-spin" />}
                {uploading ? 'Processing...' : 'Upload & Process'}
              </button>
            </div>
          )}

          {result && (
            <div className="mt-4 p-4 bg-green-50 border border-green-200 rounded-lg">
              <p className="text-sm font-medium text-green-800">{result.message}</p>
              <p className="text-xs text-green-600 mt-1">Invoice ID: {result.invoice_id}</p>
            </div>
          )}
        </div>

        {/* JSON Upload (Hackathon Demo) */}
        <div className="bg-white rounded-xl border border-gray-200 shadow-sm p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">JSON Payload (Demo)</h2>
          <textarea
            className="w-full h-64 p-4 border border-gray-300 rounded-lg font-mono text-sm focus:ring-2 focus:ring-primary-500 focus:border-transparent outline-none resize-none"
            placeholder={`{
  "vendor_name": "Acme Corp",
  "invoice_number": "INV-2026-050",
  "po_reference": "PO-2026-100",
  "total_amount": 50000,
  "tax_amount": 9000,
  "line_items": [...]
}`}
          />
          <button className="mt-4 w-full py-2.5 bg-gray-800 text-white rounded-lg text-sm font-medium hover:bg-gray-900 transition-colors">
            Submit JSON Invoice
          </button>
        </div>
      </div>
    </div>
  )
}
