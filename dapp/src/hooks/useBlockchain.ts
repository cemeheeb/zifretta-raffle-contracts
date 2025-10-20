import { useCallback, useEffect, useRef } from 'react';
import { BlockchainData } from '@workers';

export const useBlockchain = () => {
  const workerRef = useRef<Worker>();

  useEffect(() => {
    workerRef.current = new Worker(new URL('../workers/blockchain.ts', import.meta.url), {
      type: 'module'
    });

    return () => {
      workerRef.current?.terminate();
    };
  }, []);

  const fetchBlockchainData = useCallback((oracleAddress: string, userAddress: string): Promise<BlockchainData> => {
    return new Promise((resolve, reject) => {
      if (!workerRef.current) return reject('Worker not initialized');

      const handleMessage = (e: MessageEvent) => {
        if (e.data.type === 'success') {
          resolve(e.data.data.data);
          workerRef.current?.removeEventListener('message', handleMessage);
        } else if (e.data.type === 'error') {
          reject(e.data.error);
          workerRef.current?.removeEventListener('message', handleMessage);
        }
      };

      workerRef.current.addEventListener('message', handleMessage);
      workerRef.current.postMessage({ oracleAddress, userAddress });
    });
  }, []);

  return { getBlockchainData: fetchBlockchainData };
};
