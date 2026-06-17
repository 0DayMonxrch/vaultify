import { useState, useEffect, useCallback } from 'react';

export const useRevealTimer = (secretId: string, durationSeconds = 30) => {
  const [timeLeft, setTimeLeft] = useState<number>(0);

  const startTimer = useCallback(() => {
    setTimeLeft(durationSeconds);
  }, [durationSeconds]);

  const clearTimer = useCallback(() => {
    setTimeLeft(0);
  }, []);

  useEffect(() => {
    if (timeLeft <= 0) return;

    const timerId = setTimeout(() => {
      setTimeLeft(prev => prev - 1);
    }, 1000);

    return () => clearTimeout(timerId);
  }, [timeLeft]);

  // Clean up on row identity change (secretId change) to prevent dangling timers
  useEffect(() => {
    return () => clearTimer();
  }, [secretId, clearTimer]);

  return {
    timeLeft,
    startTimer,
    clearTimer,
    isRevealed: timeLeft > 0,
  };
};
