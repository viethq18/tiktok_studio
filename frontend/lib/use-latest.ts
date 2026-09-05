'use client';

import * as React from 'react';

/**
 * Keeps a ref pointing at the newest value without writing to it during render.
 * Used where a long-lived callback (the Fabric adapter, the autosave timer)
 * needs current state but must not be re-created on every render.
 */
export function useLatest<T>(value: T): React.RefObject<T> {
  const ref = React.useRef(value);
  React.useEffect(() => {
    ref.current = value;
  });
  return ref as React.RefObject<T>;
}
