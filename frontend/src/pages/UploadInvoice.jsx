import { useState, useRef } from 'react'
import { Upload, FileText, Loader2, CheckCircle, XCircle } from 'lucide-react'
import { uploadInvoice } from '../services/api'

export default function UploadInvoice() {
  const [uploading, setUploading] = useState(false)
  const [result, setResult] = useState(null)
  const [dragOver, setDragOver] = useState(false)
  const fileInputRef = useRef(null)

  const handleDrop = (e) => {
    e.preventDefault()
    setDragOver(false)
    const dropped = e.dataTransfer.files[0]
    if (dropped && dropped.type === 'application/pdf') {
      handleFileUpload(dropped)
    }
  }

  const handleFileSelect = (e) => {
    const file = e.target.files[0]
    if (file) {
      handleFileUpload(file)
    }
  }

  const handleFileUpload = async (file) => {
    setUploading(true)
    setResult(null)

    try {
      const formData = new FormData()
      formData.append('invoice', file)

      const response = await uploadInvoice(formData)

      if (response.data.success) {
        setResult({
          success: true,
          message: response.data.message || 'Invoice uploaded and processed successfully',
          invoice_id: response.data.data?.invoice_id,
          po_reference: response.data.data?.po_reference,
          invoice_number: response.data.data?.invoice_number,
          total_amount: response.data.data?.total_amount,
          currency: response.data.data?.currency,
        })
      } else {
        setResult({
          success: false,
          message: response.data.error || 'Upload failed',
        })
      }
    } catch (error) {
      const errorMessage = error.response?.data?.error || error.message || 'Upload failed'
      setResult({
        success: false,
        message: errorMessage,
      })
    } finally {
      setUploading(false)
      if (fileInputRef.current) {
        fileInputRef.current.value = ''
      }
    }
  }

  return (
    <div>
      <div className="mb-8">
        <h1 className="text-2xl font-bold text-gray-900">Upload Invoice</h1>
        <p className="text-gray-500 mt-1">Upload a PDF invoice for automatic processing</p>
      </div>

      <div className="max-w-xl mx-auto">
        <div className="bg-white rounded-xl border border-gray-200 shadow-sm p-8">
          <div
            onDragOver={(e) => { e.preventDefault(); setDragOver(true) }}
            onDragLeave={() => setDragOver(false)}
            onDrop={handleDrop}
            className={`border-2 border-dashed rounded-xl p-12 text-center transition-colors ${
              dragOver ? 'border-primary-400 bg-primary-50' : 'border-gray-300 hover:border-gray-400'
            } ${uploading ? 'pointer-events-none opacity-50' : ''}`}
          >
            {uploading ? (
              <>
                <Loader2 className="w-12 h-12 text-primary-600 mx-auto mb-4 animate-spin" />
                <p className="text-gray-600 mb-2">Processing invoice...</p>
                <p className="text-sm text-gray-400">Extracting data and validating</p>
              </>
            ) : (
              <>
                <Upload className="w-12 h-12 text-gray-400 mx-auto mb-4" />
                <p className="text-gray-600 mb-2">Drag & drop your invoice PDF here</p>
                <p className="text-sm text-gray-400 mb-4">or</p>
                <label className="px-6 py-3 bg-primary-600 text-white rounded-lg text-sm font-medium cursor-pointer hover:bg-primary-700 transition-colors inline-block">
                  Upload Invoice
                  <input
                    ref={fileInputRef}
                    type="file"
                    accept=".pdf"
                    className="hidden"
                    onChange={handleFileSelect}
                  />
                </label>
              </>
            )}
          </div>

          {result && (
            <div className={`mt-6 p-4 rounded-lg border ${
              result.success
                ? 'bg-green-50 border-green-200'
                : 'bg-red-50 border-red-200'
            }`}>
              <div className="flex items-start gap-3">
                {result.success ? (
                  <CheckCircle className="w-5 h-5 text-green-600 mt-0.5" />
                ) : (
                  <XCircle className="w-5 h-5 text-red-600 mt-0.5" />
                )}
                <div>
                  <p className={`text-sm font-medium ${
                    result.success ? 'text-green-800' : 'text-red-800'
                  }`}>
                    {result.message}
                  </p>
                  {result.success && result.invoice_id && (
                    <div className="text-xs text-green-600 mt-2 space-y-1">
                      <p>Invoice ID: {result.invoice_id}</p>
                      {result.po_reference && <p>PO Reference: {result.po_reference}</p>}
                      {result.invoice_number && <p>Invoice #: {result.invoice_number}</p>}
                      {result.total_amount && <p>Amount: {result.currency || 'INR'} {result.total_amount.toLocaleString()}</p>}
                    </div>
                  )}
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
