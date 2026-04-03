"use client"

import * as React from "react"
import { LayoutGrid, List } from "lucide-react"
import { ToggleGroup, ToggleGroupItem } from "@web/components/ui/toggle-group"
import { cn } from "@/shared/utils"

export type ViewMode = "grid" | "list"

export interface ViewToggleProps {
  /**
   * 当前视图模式
   */
  value: ViewMode
  /**
   * 视图模式变化回调
   */
  onValueChange: (value: ViewMode) => void
  /**
   * 额外的样式类名
   */
  className?: string
  /**
   * 是否禁用
   */
  disabled?: boolean
}

const ViewToggle = React.forwardRef<
  React.ElementRef<typeof ToggleGroup>,
  ViewToggleProps
>(({ value, onValueChange, className, disabled, ...props }, ref) => {
  return (
    <ToggleGroup
      ref={ref}
      type="single"
      value={value}
      onValueChange={(newValue) => newValue && onValueChange(newValue as ViewMode)}
      className={cn(className)}
      disabled={disabled}
      {...props}
    >
      <ToggleGroupItem 
        value="grid" 
        className="data-[state=on]:bg-selected data-[state=on]:text-selected-foreground"
        aria-label="网格视图"
      >
        <LayoutGrid className="size-4" />
      </ToggleGroupItem>
      <ToggleGroupItem 
        value="list" 
        className="data-[state=on]:bg-selected data-[state=on]:text-selected-foreground"
        aria-label="列表视图"
      >
        <List className="size-4" />
      </ToggleGroupItem>
    </ToggleGroup>
  )
})

ViewToggle.displayName = "ViewToggle"

export { ViewToggle }