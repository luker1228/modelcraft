import { X, Eye, Save } from 'lucide-react'

export function ButtonGroup() {
  return (
    <div className="flex gap-2">
      {/* btn-ghost: transparent bg, px-3, text-gray-900 */}
      <button
        className="inline-flex items-center gap-2 h-9 px-3 rounded-[6px] text-sm font-medium text-gray-900 bg-transparent border-none cursor-pointer transition-all duration-200 hover:bg-[#fafafa]"
      >
        <X className="w-4 h-4" strokeWidth={1.5} />
        取消
      </button>

      {/* btn-secondary: bg-[#fafafa], border border-gray-200, px-4 */}
      <button
        className="inline-flex items-center gap-2 h-9 px-4 rounded-[6px] text-sm font-medium text-gray-900 bg-[#fafafa] border border-gray-200 cursor-pointer transition-all duration-200 hover:bg-white hover:border-[#d1d5db]"
      >
        <Eye className="w-4 h-4" strokeWidth={1.5} />
        预览
      </button>

      {/* btn-primary: bg-[#2563eb], text-white, px-4 */}
      <button
        className="inline-flex items-center gap-2 h-9 px-4 rounded-[6px] text-sm font-medium text-white bg-[#2563eb] border-none cursor-pointer transition-all duration-200 hover:bg-[#1d4ed8]"
      >
        <Save className="w-4 h-4" strokeWidth={1.5} />
        保存更改
      </button>
    </div>
  )
}
