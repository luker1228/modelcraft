"use client"

import * as React from "react"
import { Search, X } from "lucide-react"
import { cva, type VariantProps } from "class-variance-authority"
import { cn } from "@/shared/utils"

const searchInputVariants = cva(
  "flex w-full rounded-md border border-input bg-background px-3 py-2 pl-10 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50",
  {
    variants: {
      size: {
        default: "h-10",
        sm: "h-8 text-xs",
        lg: "h-12",
      },
    },
    defaultVariants: {
      size: "default",
    },
  }
)

export interface SearchInputProps
  extends Omit<React.InputHTMLAttributes<HTMLInputElement>, 'size'>,
    VariantProps<typeof searchInputVariants> {
  /**
   * 是否显示清除按钮
   */
  clearable?: boolean
  /**
   * 清除按钮点击回调
   */
  onClear?: () => void
  /**
   * 容器的额外样式类名
   */
  containerClassName?: string
}

const SearchInput = React.forwardRef<HTMLInputElement, SearchInputProps>(
  ({ 
    className, 
    containerClassName,
    size, 
    clearable = false, 
    onClear, 
    value,
    ...props 
  }, ref) => {
    const showClearButton = clearable && value && String(value).length > 0

    return (
      <div className={cn("relative", containerClassName)}>
        <Search className={cn(
          "absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground",
          size === "sm" ? "w-3 h-3" : size === "lg" ? "w-5 h-5" : "w-4 h-4"
        )} />
        <input
          className={cn(searchInputVariants({ size, className }))}
          ref={ref}
          value={value}
          {...props}
        />
        {showClearButton && (
          <button
            type="button"
            onClick={onClear}
            className={cn(
              "absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground transition-colors",
              size === "sm" ? "w-3 h-3" : size === "lg" ? "w-5 h-5" : "w-4 h-4"
            )}
          >
            <X className="size-full" />
            <span className="sr-only">清除搜索</span>
          </button>
        )}
      </div>
    )
  }
)
SearchInput.displayName = "SearchInput"

export { SearchInput, searchInputVariants }