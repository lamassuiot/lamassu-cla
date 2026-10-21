import { type ReactNode, useId, useLayoutEffect, useRef, useState } from 'react';
import styles from './table.module.css';

export interface TableColumn<Row> {
  key: string;
  header: string;
  cell: (row: Row) => ReactNode;
  // Identifies the row, for example a name; rendered as a row header.
  rowHeader?: boolean;
  numeric?: boolean;
}

export interface TableProps<Row> {
  // The accessible name of the table and of its scroll container.
  caption: string;
  columns: TableColumn<Row>[];
  rows: Row[];
  rowKey: (row: Row) => string;
}

// Only tables may scroll horizontally (AC-05-11). When the table overflows, its container becomes a
// tab stop so keyboard users can scroll it.
export function Table<Row>({ caption, columns, rows, rowKey }: TableProps<Row>) {
  const captionId = useId();
  const scrollRef = useRef<HTMLElement>(null);
  const [scrollable, setScrollable] = useState(false);

  useLayoutEffect(() => {
    const container = scrollRef.current;
    if (!container) {
      return;
    }
    const update = () => setScrollable(container.scrollWidth > container.clientWidth);
    update();
    const observer = new ResizeObserver(update);
    observer.observe(container);
    for (const child of container.children) {
      observer.observe(child);
    }
    return () => observer.disconnect();
  }, []);

  return (
    <section
      ref={scrollRef}
      className={styles.scroll}
      aria-labelledby={captionId}
      tabIndex={scrollable ? 0 : undefined}
    >
      <table className={styles.table}>
        <caption id={captionId} className={styles.caption}>
          {caption}
        </caption>
        <thead>
          <tr>
            {columns.map((column) => (
              <th
                key={column.key}
                scope="col"
                className={column.numeric ? styles.numeric : undefined}
              >
                {column.header}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {rows.map((row) => (
            <tr key={rowKey(row)}>
              {columns.map((column) => {
                const Cell = column.rowHeader ? 'th' : 'td';
                return (
                  <Cell
                    key={column.key}
                    scope={column.rowHeader ? 'row' : undefined}
                    className={column.numeric ? styles.numeric : undefined}
                  >
                    {column.cell(row)}
                  </Cell>
                );
              })}
            </tr>
          ))}
        </tbody>
      </table>
    </section>
  );
}
