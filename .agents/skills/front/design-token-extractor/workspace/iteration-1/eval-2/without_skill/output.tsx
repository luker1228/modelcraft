import { LayoutDashboard, Database, Settings } from 'lucide-react'

export default function Sidebar() {
  return (
    <div className="sidebar">
      <div className="sidebar-content">
        <div className="sidebar-section-label">工作区</div>
        <a className="sidebar-nav-item active" href="#">
          <LayoutDashboard width={16} height={16} />
          总览
        </a>
        <a className="sidebar-nav-item" href="#">
          <Database width={16} height={16} />
          数据模型
        </a>
        <a className="sidebar-nav-item" href="#">
          <Settings width={16} height={16} />
          设置
        </a>
      </div>
    </div>
  )
}
