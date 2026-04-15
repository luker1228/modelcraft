'use client'

import { PlusCircle, Settings, Tags } from 'lucide-react'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from '@web/components/ui/sheet'
import type { ModelEditorState } from '../_hooks'
import type { FieldOperations } from '../_hooks'
import {
  CreateEnumFieldPage,
  CreateEnumLabelFieldPage,
  EditFieldImmutablePage,
} from './field-pages'

interface FieldEditSheetProps {
  state: ModelEditorState
  fieldOps: FieldOperations
  orgName: string
  projectSlug: string
}

export function FieldEditSheet({ state, fieldOps, orgName, projectSlug }: FieldEditSheetProps) {
  const pageMeta =
    fieldOps.fieldPageMode === 'create-enum'
      ? {
          title: '创建 ENUM 字段',
          description: '填写基础信息并绑定枚举。',
          icon: <Tags className="size-4 text-primary" />,
        }
      : fieldOps.fieldPageMode === 'create-enum-label'
        ? {
            title: '创建 ENUM_LABEL 字段',
            description: '选择 source 字段，保存时自动创建并绑定 relation。',
            icon: <PlusCircle className="size-4 text-primary" />,
          }
        : {
            title: '编辑字段',
            description: `编辑字段 ${state.editingField?.name || ''}，format 与关联配置只读。`,
            icon: <Settings className="size-4 text-primary" />,
          }

  return (
    <Sheet open={state.editFieldOpen} onOpenChange={state.setEditFieldOpen}>
      <SheetContent className="w-[520px] overflow-y-auto sm:max-w-[520px]">
        <SheetHeader>
          <SheetTitle className="flex items-center gap-2 text-base">
            {pageMeta.icon}
            {pageMeta.title}
          </SheetTitle>
          <SheetDescription className="text-sm">{pageMeta.description}</SheetDescription>
        </SheetHeader>

        {fieldOps.fieldPageMode === 'create-enum' && (
          <CreateEnumFieldPage
            enumOptions={fieldOps.enumOptions}
            loading={fieldOps.contextLoading || fieldOps.createEnumFieldLoading}
            error={fieldOps.createEnumFieldError ?? fieldOps.contextError}
            onSubmit={fieldOps.handleSubmitCreateEnumField}
            onCancel={fieldOps.handleCloseFieldPage}
            orgName={orgName}
            projectSlug={projectSlug}
          />
        )}

        {fieldOps.fieldPageMode === 'create-enum-label' && (
          <CreateEnumLabelFieldPage
            sourceOptions={fieldOps.sourceOptions}
            loading={fieldOps.contextLoading || fieldOps.createEnumLabelFieldLoading}
            error={fieldOps.createEnumLabelFieldError ?? fieldOps.contextError}
            onSubmit={fieldOps.handleSubmitCreateEnumLabelField}
            onCancel={fieldOps.handleCloseFieldPage}
          />
        )}

        {fieldOps.fieldPageMode === 'edit' && state.editingField && (
          <EditFieldImmutablePage
            fieldName={state.editingField.name}
            title={state.editingField.title}
            description={state.editingField.description}
            format={state.editingField.format || '-'}
            relateEnumName={state.editingField.enum?.name}
            enumRelationId={state.editingField.enumRelationId}
            loading={fieldOps.editFieldLoading}
            error={fieldOps.editFieldError}
            onSubmit={fieldOps.handleSubmitEditField}
            onCancel={fieldOps.handleCloseFieldPage}
          />
        )}
      </SheetContent>
    </Sheet>
  )
}
