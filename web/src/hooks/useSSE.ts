import { useEffect, useRef, useCallback } from 'react';

/**
 * useSSE subscribes to the server-sent events stream for real-time updates.
 * Events are dispatched to the provided callback.
 */
export function useSSE(onEvent: (event: MessageEvent) => void) {
  const eventSourceRef = useRef<EventSource | null>(null);
  const callbackRef = useRef(onEvent);
  callbackRef.current = onEvent;

  const connect = useCallback(() => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
    }

    const es = new EventSource('/api/v1/events');

    es.onmessage = (event) => {
      callbackRef.current(event);
    };

    es.addEventListener('connected', () => {
      console.log('SSE connected');
    });

    es.addEventListener('asset.created', (event) => {
      callbackRef.current(event);
    });

    es.addEventListener('asset.updated', (event) => {
      callbackRef.current(event);
    });

    es.addEventListener('asset.removed', (event) => {
      callbackRef.current(event);
    });

    es.onerror = () => {
      console.warn('SSE connection error, will reconnect...');
      es.close();
      setTimeout(connect, 5000);
    };

    eventSourceRef.current = es;
  }, []);

  useEffect(() => {
    connect();
    return () => {
      eventSourceRef.current?.close();
    };
  }, [connect]);
}
