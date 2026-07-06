import { useState, useRef, useCallback } from 'react';
import { convertToWav } from '../utils/wavConverter';

export function useAudioRecorder() {
    const [isRecording, setIsRecording] = useState(false);
    const [recordingProgress, setRecordingProgress] = useState(0);
    const [recordedWav, setRecordedWav] = useState<Blob | null>(null);
    const [error, setError] = useState<string | null>(null);

    const mediaRecorder = useRef<MediaRecorder | null>(null);
    const audioChunks = useRef<Blob[]>([]);
    const progressInterval = useRef<number | null>(null);

    const startRecording = useCallback(async () => {
        setError(null); setRecordedWav(null); setRecordingProgress(0);
        audioChunks.current = [];

        try {
            const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
            mediaRecorder.current = new MediaRecorder(stream);
            mediaRecorder.current.ondataavailable = (e) => { if (e.data.size > 0) audioChunks.current.push(e.data); };
            mediaRecorder.current.onstop = async () => {
                setIsRecording(false);
                if (progressInterval.current) clearInterval(progressInterval.current);
                stream.getTracks().forEach(t => t.stop());
                const rawBlob = new Blob(audioChunks.current, { type: mediaRecorder.current?.mimeType || 'audio/webm' });
                try {
                    const wavBlob = await convertToWav(rawBlob);
                    setRecordedWav(wavBlob);
                } catch (err) {
                    setError('Failed to process format.');
                }
            };

            mediaRecorder.current.start();
            setIsRecording(true);

            const startTime = Date.now();
            progressInterval.current = window.setInterval(() => {
                const elapsed = Date.now() - startTime;
                setRecordingProgress(Math.min((elapsed / 5000) * 100, 100));
                if (elapsed >= 5000 && mediaRecorder.current?.state === 'recording') mediaRecorder.current.stop();
            }, 50);
        } catch (err: any) {
            setError('Microphone access denied.');
        }
    }, []);

    return { isRecording, recordingProgress, startRecording, recordedWav, error };
}