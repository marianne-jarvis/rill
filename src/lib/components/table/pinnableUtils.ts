import type { ValidationState } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";

export interface ColumnConfig {
  name: string;
  type: string;

  renderer?: unknown;

  validation?: (
    row: Record<string, unknown>,
    value: unknown
  ) => ValidationState;

  copyable?: boolean;
}

export function columnIsPinned(name, selectedColumns: Array<ColumnConfig>) {
  return selectedColumns.map((column) => column.name).includes(name);
}

export function togglePin(name, type, selectedColumns) {
  // if column is already pinned, remove.
  if (columnIsPinned(name, selectedColumns)) {
    return [...selectedColumns.filter((column) => column.name !== name)];
  } else {
    return [...selectedColumns, { name, type }];
  }
}
