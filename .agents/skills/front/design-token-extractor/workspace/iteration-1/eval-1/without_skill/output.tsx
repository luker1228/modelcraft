import { Info } from 'lucide-react'

export default function BasicInfoCard() {
  return (
    <div className="bg-white border border-gray-200 rounded-lg shadow-sm">
      <div className="flex items-center gap-2 px-6 py-4 border-b border-gray-200">
        <Info className="w-5 h-5 text-gray-500" strokeWidth={1.5} />
        <span className="font-semibold text-gray-900">基本信息</span>
      </div>
      <div className="px-6 py-4 space-y-4">
        <div className="flex items-start gap-4">
          <div className="w-32 shrink-0">
            <div className="text-sm font-medium text-gray-700">
              显示名称<span className="text-red-500 ml-0.5">*</span>
            </div>
          </div>
          <div className="flex-1">
            <input
              type="text"
              className="w-full px-3 py-2 text-sm border border-gray-200 rounded-md bg-white text-gray-900 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              defaultValue="电商数据库"
            />
          </div>
        </div>
        <div className="flex items-start gap-4">
          <div className="w-32 shrink-0">
            <div className="text-sm font-medium text-gray-700">描述</div>
            <div className="text-xs text-gray-400 mt-0.5">输入简短描述</div>
          </div>
          <div className="flex-1">
            <input
              type="text"
              className="w-full px-3 py-2 text-sm border border-gray-200 rounded-md bg-white text-gray-900 placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              placeholder="输入描述..."
            />
          </div>
        </div>
      </div>
      <div className="flex items-center justify-end gap-2 px-6 py-4 border-t border-gray-200">
        <button className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-200 rounded-md hover:bg-gray-50 transition-colors">
          取消
        </button>
        <button className="px-4 py-2 text-sm font-medium text-white bg-[#2563eb] rounded-md hover:bg-blue-700 transition-colors">
          保存
        </button>
      </div>
    </div>
  )
}
