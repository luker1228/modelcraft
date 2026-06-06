// @vitest-environment jsdom
import '@testing-library/jest-dom/vitest'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { ModelRecordTable, type ModelRecordTableFieldInfo } from './ModelRecordTable'

afterEach(cleanup)

const fieldInfo: ModelRecordTableFieldInfo = {
  name: 'name',
  title: 'Name',
  isPrimary: false,
  isDeprecated: false,
  storageHint: 'TEXT',
  schemaType: 'STRING',
}

describe('ModelRecordTable pagination', () => {
  it('renders pagination controls and applies row number offset', () => {
    const onPreviousPage = vi.fn()
    const onNextPage = vi.fn()

    render(
      <ModelRecordTable
        contentLoading={false}
        contentList={[
          { id: 'p021', name: 'Row 21' },
          { id: 'p022', name: 'Row 22' },
        ]}
        displayFields={['name']}
        getFieldInfo={() => fieldInfo}
        getFieldTypeDisplay={() => 'TEXT'}
        propByName={{}}
        onCreate={vi.fn()}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        pagination={{
          currentPage: 2,
          pageSize: 20,
          hasNextPage: true,
          onPreviousPage,
          onNextPage,
        }}
      />,
    )

    expect(screen.getByText('21')).toBeInTheDocument()
    expect(screen.getByText('22')).toBeInTheDocument()
    expect(screen.getByText('第 2 页')).toBeInTheDocument()

    fireEvent.click(screen.getByRole('button', { name: '上一页' }))
    fireEvent.click(screen.getByRole('button', { name: '下一页' }))

    expect(onPreviousPage).toHaveBeenCalledTimes(1)
    expect(onNextPage).toHaveBeenCalledTimes(1)
  })

  it('disables previous button on the first page and next button when no more pages', () => {
    render(
      <ModelRecordTable
        contentLoading={false}
        contentList={[{ id: 'p001', name: 'Row 1' }]}
        displayFields={['name']}
        getFieldInfo={() => fieldInfo}
        getFieldTypeDisplay={() => 'TEXT'}
        propByName={{}}
        onCreate={vi.fn()}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        pagination={{
          currentPage: 1,
          pageSize: 20,
          hasNextPage: false,
          onPreviousPage: vi.fn(),
          onNextPage: vi.fn(),
        }}
      />,
    )

    expect(screen.getByRole('button', { name: '上一页' })).toBeDisabled()
    expect(screen.getByRole('button', { name: '下一页' })).toBeDisabled()
  })
})
