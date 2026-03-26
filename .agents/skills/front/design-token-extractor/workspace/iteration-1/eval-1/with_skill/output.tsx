import { Info } from 'lucide-react'

interface BasicInfoCardProps {
  displayName?: string
  description?: string
  onCancel?: () => void
  onSave?: () => void
}

export default function BasicInfoCard({
  displayName = '电商数据库',
  description = '',
  onCancel,
  onSave,
}: BasicInfoCardProps) {
  return (
    <div className="bg-white border border-gray-200 rounded-lg overflow-hidden">
      {/* Card Header */}
      <div className="flex items-center gap-2 px-5 py-4 border-b border-gray-200">
        <Info className="w-4 h-4 text-gray-500" strokeWidth={1.5} />
        <span className="text-sm font-semibold text-gray-900">基本信息</span>
      </div>

      {/* Card Body */}
      <div className="p-5">
        {/* Form Row: 显示名称 */}
        <div className="flex items-start gap-6 mb-6">
          <div className="w-[38%] flex-none">
            <div className="text-sm font-medium text-gray-700">
              显示名称<span className="text-red-500">*</span>
            </div>
          </div>
          <div className="flex-1">
            <input
              type="text"
              className="w-full h-[34px] px-[10px] text-[13px] border border-gray-200 rounded-[6px] bg-white text-gray-900 placeholder-gray-400 focus:outline-none focus:border-[#2563eb] focus:ring-2 focus:ring-[#2563eb]/10"
              defaultValue={displayName}
            />
          </div>
        </div>

        {/* Form Row: 描述 */}
        <div className="flex items-start gap-6 mb-6">
          <div className="w-[38%] flex-none">
            <div className="text-sm font-medium text-gray-700">描述</div>
            <div className="text-xs text-gray-400 mt-1">输入简短描述</div>
          </div>
          <div className="flex-1">
            <input
              type="text"
              className="w-full h-[34px] px-[10px] text-[13px] border border-gray-200 rounded-[6px] bg-white text-gray-900 placeholder-gray-400 focus:outline-none focus:border-[#2563eb] focus:ring-2 focus:ring-[#2563eb]/10"
              placeholder="输入描述..."
              defaultValue={description}
            />
          </div>
        </div>
      </div>

      {/* Card Footer */}
      <div className="flex justify-end gap-2 py-3 px-5 border-t border-gray-200 bg-gray-50">
        <button
          onClick={onCancel}
          className="inline-flex items-center h-9 px-4 text-sm font-medium text-gray-900 bg-gray-50 border border-gray-200 rounded-[6px] cursor-pointer hover:bg-white hover:border-[#d1d5db] transition-all duration-200"
        >
          取消
        </button>
        <button
          onClick={onSave}
          className="inline-flex items-center h-9 px-4 text-sm font-medium text-white bg-[#2563eb] border-none rounded-[6px] cursor-pointer hover:bg-[#1d4ed8] transition-all duration-200"
        >
          保存
        </button>
      </div>
    </div>
  )
}
