import { useState, useCallback } from 'react'
import { useAppStore } from '@web/stores/app'
import type { EditorModel, EditorModelDetail, EditorModelField } from './types'
import type { LogicalForeignKey } from '@/types'

export function useModelEditorState() {
  // Database / search / selection
  const storedDatabase = useAppStore((s) => s.selectedDatabase)
  const setStoredDatabase = useAppStore((s) => s.setSelectedDatabase)
  const [selectedDatabase, _setSelectedDatabase] = useState<string>(storedDatabase ?? '')
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedModelId, setSelectedModelId] = useState<string | null>(null)
  const [databaseOpen, setDatabaseOpen] = useState(false)

  const setSelectedDatabase = useCallback(
    (db: string) => {
      _setSelectedDatabase(db)
      setStoredDatabase(db)
    },
    [setStoredDatabase],
  )

  // Connection status
  const [connectionChecking, setConnectionChecking] = useState(true)
  const [connectionFailed, setConnectionFailed] = useState(false)
  const [connectionError, setConnectionError] = useState<string>('')

  // Create model dialog
  const [createModelOpen, setCreateModelOpen] = useState(false)
  const [newModelName, setNewModelName] = useState('')
  const [newModelTitle, setNewModelTitle] = useState('')
  const [creating, setCreating] = useState(false)
  const [importDialogOpen, setImportDialogOpen] = useState(false)

  // Edit model state
  const [editModelOpen, setEditModelOpen] = useState(false)
  const [editModelId, setEditModelId] = useState<string | null>(null)
  const [editModelData, setEditModelData] = useState<EditorModelDetail | null>(null)
  const [editModelLoading, setEditModelLoading] = useState(false)

  // Delete model confirmation
  const [deleteModelDialogOpen, setDeleteModelDialogOpen] = useState(false)
  const [modelToDelete, setModelToDelete] = useState<EditorModel | null>(null)
  const [deletingModel, setDeletingModel] = useState(false)

  // Meta info inline editing
  const [metaTitle, setMetaTitle] = useState('')
  const [metaDescription, setMetaDescription] = useState('')
  const [metaDisplayField, setMetaDisplayField] = useState('')
  const [metaInsertionOrderField, setMetaInsertionOrderField] = useState('')
  const [metaSaving, setMetaSaving] = useState(false)
  const [metaEditMode, setMetaEditMode] = useState(false)

  // Insert field
  const [insertFieldOpen, setInsertFieldOpen] = useState(false)

  // Edit field
  const [editFieldOpen, setEditFieldOpen] = useState(false)
  const [editingField, setEditingField] = useState<EditorModelField | null>(null)
  const [editFieldTitle, setEditFieldTitle] = useState('')
  const [editFieldDescription, setEditFieldDescription] = useState('')

  // Foreign keys
  const [fkList, setFkList] = useState<LogicalForeignKey[]>([])
  const [fkLoading, setFkLoading] = useState(false)
  const [fkFormOpen, setFkFormOpen] = useState(false)
  const [fkRefModelId, setFkRefModelId] = useState('')
  const [fkRefModelDetail, setFkRefModelDetail] = useState<EditorModelDetail | null>(null)
  const [fkRefModelLoading, setFkRefModelLoading] = useState(false)
  const [fkMappings, setFkMappings] = useState<{ sourceField: string; targetField: string }[]>([
    { sourceField: '', targetField: '' },
  ])
  const [fkSubmitting, setFkSubmitting] = useState(false)
  const [fkDeleteConfirm, setFkDeleteConfirm] = useState<string | null>(null)

  return {
    // Database / search / selection
    selectedDatabase, setSelectedDatabase,
    searchQuery, setSearchQuery,
    selectedModelId, setSelectedModelId,
    databaseOpen, setDatabaseOpen,

    // Connection
    connectionChecking, setConnectionChecking,
    connectionFailed, setConnectionFailed,
    connectionError, setConnectionError,

    // Create model
    createModelOpen, setCreateModelOpen,
    newModelName, setNewModelName,
    newModelTitle, setNewModelTitle,
    creating, setCreating,
    importDialogOpen, setImportDialogOpen,

    // Edit model
    editModelOpen, setEditModelOpen,
    editModelId, setEditModelId,
    editModelData, setEditModelData,
    editModelLoading, setEditModelLoading,

    // Delete model
    deleteModelDialogOpen, setDeleteModelDialogOpen,
    modelToDelete, setModelToDelete,
    deletingModel, setDeletingModel,

    // Meta editing
    metaTitle, setMetaTitle,
    metaDescription, setMetaDescription,
    metaDisplayField, setMetaDisplayField,
    metaInsertionOrderField, setMetaInsertionOrderField,
    metaSaving, setMetaSaving,
    metaEditMode, setMetaEditMode,

    // Insert field
    insertFieldOpen, setInsertFieldOpen,

    // Edit field
    editFieldOpen, setEditFieldOpen,
    editingField, setEditingField,
    editFieldTitle, setEditFieldTitle,
    editFieldDescription, setEditFieldDescription,

    // Foreign keys
    fkList, setFkList,
    fkLoading, setFkLoading,
    fkFormOpen, setFkFormOpen,
    fkRefModelId, setFkRefModelId,
    fkRefModelDetail, setFkRefModelDetail,
    fkRefModelLoading, setFkRefModelLoading,
    fkMappings, setFkMappings,
    fkSubmitting, setFkSubmitting,
    fkDeleteConfirm, setFkDeleteConfirm,
  }
}

export type ModelEditorState = ReturnType<typeof useModelEditorState>
