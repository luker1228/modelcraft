import { X, Eye, Save } from 'lucide-react'

export default function ButtonGroup() {
  return (
    <div style={{ display: 'flex', gap: '8px' }}>
      <button className="btn-ghost">
        <X width={16} height={16} />
        取消
      </button>
      <button className="btn-secondary">
        <Eye width={16} height={16} />
        预览
      </button>
      <button className="btn-primary">
        <Save width={16} height={16} />
        保存更改
      </button>
    </div>
  )
}
