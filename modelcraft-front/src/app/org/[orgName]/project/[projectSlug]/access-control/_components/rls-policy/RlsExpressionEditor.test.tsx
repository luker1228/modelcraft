// @vitest-environment jsdom
import '@testing-library/jest-dom/vitest'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { RlsExpressionEditor } from './RlsExpressionEditor'

describe('RlsExpressionEditor', () => {
  afterEach(cleanup)

  it('inserts the clicked field at the current cursor position', () => {
    const onChange = vi.fn()

    render(
      <RlsExpressionEditor
        label="Using Filter"
        placeholder="例如：row.owner_id == auth.userid"
        example="row.owner_id == auth.userid"
        availableFields={['row.owner_id', 'row.status']}
        rootLabel="row"
        value='row.status == "active" && '
        onChange={onChange}
        exprType="SELECT_PREDICATE"
      />,
    )

    const textarea = screen.getByPlaceholderText(
      '例如：row.owner_id == auth.userid',
    ) as HTMLTextAreaElement
    fireEvent.focus(textarea)
    const insertAt = 'row.status == "active" && '.length
    textarea.setSelectionRange(insertAt, insertAt)

    fireEvent.click(screen.getByRole('button', { name: 'row.owner_id' }))

    expect(onChange).toHaveBeenCalledWith('row.status == "active" && row.owner_id')
  })

  it('adds spaces around the inserted field when surrounding text is tight', () => {
    const onChange = vi.fn()

    render(
      <RlsExpressionEditor
        label="Using Filter"
        placeholder="例如：row.owner_id == auth.userid"
        example="row.owner_id == auth.userid"
        availableFields={['row.owner_id']}
        rootLabel="row"
        value='row.status=="active"&&auth.userid=='
        onChange={onChange}
        exprType="SELECT_PREDICATE"
      />,
    )

    const textarea = screen.getByPlaceholderText(
      '例如：row.owner_id == auth.userid',
    ) as HTMLTextAreaElement
    fireEvent.focus(textarea)
    const insertAt = 'row.status=="active"&&'.length
    textarea.setSelectionRange(insertAt, insertAt)

    fireEvent.click(screen.getByRole('button', { name: 'row.owner_id' }))

    expect(onChange).toHaveBeenCalledWith(
      'row.status=="active"&& row.owner_id auth.userid==',
    )
  })

  it('replaces the selected text with the clicked field', () => {
    const onChange = vi.fn()

    render(
      <RlsExpressionEditor
        label="Using Filter"
        placeholder="例如：row.owner_id == auth.userid"
        example="row.owner_id == auth.userid"
        availableFields={['row.owner_id']}
        rootLabel="row"
        value='row.status == old_value && auth.userid == 1'
        onChange={onChange}
        exprType="SELECT_PREDICATE"
      />,
    )

    const textarea = screen.getByPlaceholderText(
      '例如：row.owner_id == auth.userid',
    ) as HTMLTextAreaElement
    fireEvent.focus(textarea)
    const start = 'row.status == '.length
    const end = start + 'old_value'.length
    textarea.setSelectionRange(start, end)

    fireEvent.click(screen.getByRole('button', { name: 'row.owner_id' }))

    expect(onChange).toHaveBeenCalledWith(
      'row.status == row.owner_id && auth.userid == 1',
    )
  })

  it('shows completion candidates for row token and replaces the current token', () => {
    const onChange = vi.fn()

    render(
      <RlsExpressionEditor
        label="Using Filter"
        placeholder="例如：row.owner_id == auth.userid"
        example="row.owner_id == auth.userid"
        availableFields={['row.owner_id', 'row.status']}
        modelFields={[
          { name: 'owner_id', title: 'Owner' },
          { name: 'status', title: 'Status' },
        ]}
        authVariables={[{ name: 'userid', type: 'string', source: 'X-MC-Auth-Userid' }]}
        rootLabel="row"
        value="row."
        onChange={onChange}
        exprType="SELECT_PREDICATE"
      />,
    )

    const textarea = screen.getByPlaceholderText(
      '例如：row.owner_id == auth.userid',
    ) as HTMLTextAreaElement
    fireEvent.focus(textarea)
    textarea.setSelectionRange('row.'.length, 'row.'.length)
    fireEvent.select(textarea)

    fireEvent.click(screen.getByRole('button', { name: 'row.owner_id 候选' }))

    expect(onChange).toHaveBeenCalledWith('row.owner_id')
  })
})
