import React, { useState, useRef } from 'react'
import {
  UploadCloud,
  FileText,
  CheckCircle2,
  FolderOpen,
} from 'lucide-react'
import { DocumentItem } from '../types'
import { useAuth } from '../context/AuthContext'

interface DocumentsViewProps {
  documents: DocumentItem[]
  onUploadFile: (file: File) => Promise<void>
}

export const DocumentsView: React.FC<DocumentsViewProps> = ({ documents, onUploadFile }) => {
  const { t } = useAuth()
  const [isDragOver, setIsDragOver] = useState(false)
  const [isUploading, setIsUploading] = useState(false)
  const fileInputRef = useRef<HTMLInputElement>(null)

  const handleDrop = async (e: React.DragEvent) => {
    e.preventDefault()
    setIsDragOver(false)

    if (e.dataTransfer.files.length > 0) {
      const file = e.dataTransfer.files[0]
      setIsUploading(true)
      try {
        await onUploadFile(file)
      } finally {
        setIsUploading(false)
      }
    }
  }

  const handleFilePick = async () => {
    // If running in Electron, try native dialog first
    if (window.claimpilotDesktop) {
      const filePath = await window.claimpilotDesktop.openFileDialog()
      if (filePath) {
        window.claimpilotDesktop.notify('Document Intake', `Selected ${filePath}`)
      }
      return
    }
    // Web fallback: trigger hidden file input
    fileInputRef.current?.click()
  }

  const handleFileInputChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = e.target.files
    if (files && files.length > 0) {
      const file = files[0]
      setIsUploading(true)
      try {
        await onUploadFile(file)
      } finally {
        setIsUploading(false)
        // Reset input so the same file can be selected again
        if (fileInputRef.current) {
          fileInputRef.current.value = ''
        }
      }
    }
  }

  return (
    <div>
      <div className="page-header">
        <div className="page-title-group">
          <h1>{t.documentsTitle}</h1>
          <p className="page-subtitle">{t.documentsSubtitle}</p>
        </div>
      </div>

      {/* Hidden file input for web fallback */}
      <input
        ref={fileInputRef}
        type="file"
        accept=".pdf,.txt,.csv,.md,.doc,.docx"
        style={{ display: 'none' }}
        onChange={handleFileInputChange}
      />

      {/* Drag & Drop Upload Zone */}
      <div
        className={`dropzone ${isDragOver ? 'dragover' : ''} ${isUploading ? 'uploading' : ''}`}
        onDragOver={(e) => {
          e.preventDefault()
          setIsDragOver(true)
        }}
        onDragLeave={() => setIsDragOver(false)}
        onDrop={handleDrop}
        onClick={handleFilePick}
      >
        <div className="dropzone-icon-wrapper">
          <UploadCloud size={36} className="dropzone-icon" />
        </div>
        <div className="dropzone-title">
          {isUploading ? t.analyzingWithAi : t.dragDropTitle}
        </div>
        <p className="dropzone-desc" style={{ marginBottom: 14 }}>
          {t.dragDropSubtitle}
        </p>
        <button className="btn btn-secondary" onClick={(e) => { e.stopPropagation(); handleFilePick() }}>
          <FolderOpen size={14} /> {t.browseFiles}
        </button>
      </div>

      {/* Uploaded Documents List */}
      <div className="table-container">
        <table>
          <thead>
            <tr>
              <th>{t.fileNameCol}</th>
              <th>{t.typeCol}</th>
              <th>{t.statusCol}</th>
              <th>{t.sizeCol}</th>
              <th>{t.summaryCol}</th>
              <th>{t.dateCol}</th>
            </tr>
          </thead>
          <tbody>
            {documents.length === 0 ? (
              <tr>
                <td colSpan={6} style={{ textAlign: 'center', padding: '32px 0', color: 'var(--text-muted)' }}>
                  Henüz belge yüklenmedi.
                </td>
              </tr>
            ) : (
              documents.map((doc) => (
                <tr key={doc.id}>
                  <td>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 8, fontWeight: 600, color: 'var(--text-primary)' }}>
                      <FileText size={15} color="#38bdf8" />
                      {doc.fileName}
                    </div>
                  </td>
                  <td>
                    <span className="badge badge-medium">{doc.fileType}</span>
                  </td>
                  <td>
                    <span className="badge badge-success" style={{ gap: 4 }}>
                      <CheckCircle2 size={11} /> {doc.status}
                    </span>
                  </td>
                  <td className="card-value mono" style={{ fontSize: 12 }}>
                    {(doc.fileSize / 1024).toFixed(1)} KB
                  </td>
                  <td style={{ maxWidth: 360, color: 'var(--text-secondary)' }}>
                    {doc.summary || 'Yükümlülük çıkarma ve vade takibi aktif.'}
                  </td>
                  <td>{new Date(doc.createdAt).toLocaleDateString()}</td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  )
}
