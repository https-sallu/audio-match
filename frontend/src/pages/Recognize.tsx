import { useState, useEffect } from 'react';
import { useAudioRecorder } from '../hooks/useAudioRecorder';
import { api, MatchResult } from '../services/api';
import { Mic, Loader, CheckCircle, Search, Volume2, AlertCircle } from 'lucide-react';

export default function Recognize() {
    // Extract hook state safely to allow manual stopping for quick/deep scans
    const recorderState = useAudioRecorder();
    const { isRecording, startRecording, recordedWav, error: micErr } = recorderState;
    const stopRecording = (recorderState as any).stopRecording;

    const [isProcessing, setIsProcessing] = useState(false);
    const [result, setResult] = useState<MatchResult | null>(null);
    const [error, setError] = useState<string | null>(null);

    // State Machine for Smart Scanning
    const [scanAttempt, setScanAttempt] = useState<0 | 1 | 2>(0);
    const [statusMsg, setStatusMsg] = useState<{ text: string, color: string, icon: any } | null>(null);

    // Track processed blobs so we don't accidentally re-process the exact same file
    const [processedWav, setProcessedWav] = useState<Blob | null>(null);

    const handleRecordClick = () => {
        setResult(null);
        setError(null);
        setScanAttempt(1); // Start Stage 1
        setProcessedWav(null); // Reset for new session
        setStatusMsg({ text: "Listening to audio (Quick Scan)...", color: "text-blue-400", icon: <Mic className="w-4 h-4" /> });
        startRecording();

        // 1. Early Match Check: Try to stop after 2.5 seconds for a quick scan
        setTimeout(() => {
            if (stopRecording) stopRecording();
        }, 2500);
    };

    useEffect(() => {
        // Only process if we have a NEW audio blob and are actively scanning
        if (recordedWav && recordedWav !== processedWav && scanAttempt > 0) {
            setProcessedWav(recordedWav); // Mark this blob as handled
            setIsProcessing(true);
            setStatusMsg({ text: "Analyzing frequencies...", color: "text-yellow-400", icon: <Search className="w-4 h-4" /> });

            api.recognizeAudio(recordedWav)
                .then((res) => {
                    // EARLY MATCH SUCCESS!
                    setResult(res);
                    setScanAttempt(0); // Reset machine
                    setStatusMsg(null);
                })
                .catch((e) => {
                    if (scanAttempt === 1) {
                        // STAGE 1 FAILED: Trigger the Extended Deep Scan (Stage 2)
                        setScanAttempt(2);
                        setStatusMsg({
                            text: "Move closer or play it louder... Extending scan 📡",
                            color: "text-orange-400",
                            icon: <Volume2 className="w-5 h-5 animate-bounce" />
                        });

                        // Give the user 2 seconds to read the message and move their phone
                        setTimeout(() => {
                            setStatusMsg({ text: "Listening to audio (Deep Scan)...", color: "text-orange-400", icon: <Mic className="w-4 h-4" /> });
                            startRecording();

                            // 2. Extended Check: Record for 7.5s (making 10s total)
                            setTimeout(() => {
                                if (stopRecording) stopRecording();
                            }, 7500);
                        }, 2000);

                    } else {
                        // STAGE 2 FAILED: Give up after 10 total seconds.
                        setError("No match found in the database. Please try again.");
                        setScanAttempt(0); // Reset machine
                        setStatusMsg(null);
                    }
                })
                .finally(() => {
                    setIsProcessing(false);
                });
        }
    }, [recordedWav, scanAttempt, processedWav, stopRecording, startRecording]);

    // Handle hardware microphone errors
    useEffect(() => {
        if (micErr) {
            setError(micErr);
            setScanAttempt(0);
            setStatusMsg(null);
        }
    }, [micErr]);

    return (
        <div className="flex flex-col items-center justify-center min-h-[60vh] space-y-12">
            <div className="text-center space-y-4">
                <h2 className="text-4xl font-extrabold text-white tracking-tight">Tap to Shazam</h2>
                <p className="text-gray-400">Discover the music playing around you</p>
            </div>

            <div className="relative group flex flex-col items-center">
                {/* Radar Ripple Animation Background (Changes color on Extended Scan) */}
                {isRecording && (
                    <div className={`absolute inset-0 rounded-full animate-ping opacity-75 scale-150 ${scanAttempt === 2 ? 'bg-orange-500' : 'bg-blue-500'}`}></div>
                )}

                <button
                    onClick={handleRecordClick}
                    disabled={isRecording || isProcessing || scanAttempt > 0}
                    className={`relative z-10 p-10 rounded-full flex items-center justify-center shadow-2xl transition-all duration-300 ${isRecording
                        ? (scanAttempt === 2 ? 'bg-orange-600 scale-110 shadow-orange-500/50' : 'bg-red-500 scale-110 shadow-red-500/50')
                        : 'bg-blue-600 hover:bg-blue-500 hover:scale-105 shadow-blue-900/50'
                        }`}
                >
                    {isProcessing ? (
                        <Loader className="w-16 h-16 text-white animate-spin" />
                    ) : (
                        <Mic className={`w-16 h-16 text-white ${isRecording ? 'animate-pulse' : ''}`} />
                    )}
                </button>

                {/* Dynamic Status Text */}
                <div className="mt-8 h-8 flex items-center justify-center">
                    {(isRecording || isProcessing || scanAttempt > 0) && statusMsg && (
                        <p className={`${statusMsg.color} font-medium animate-pulse flex items-center gap-2 text-lg`}>
                            {statusMsg.icon} {statusMsg.text}
                        </p>
                    )}
                </div>
            </div>

            {/* Modern Error Box */}
            {(error) && scanAttempt === 0 && !isRecording && !isProcessing && (
                <div className="text-red-400 bg-red-900/20 px-6 py-4 rounded-xl border border-red-900/50 shadow-lg flex items-center gap-3 animate-in fade-in slide-in-from-bottom-4">
                    <AlertCircle className="w-6 h-6" />
                    <p className="font-medium">{error}</p>
                </div>
            )}

            {/* Success Box */}
            {result && scanAttempt === 0 && !isRecording && !isProcessing && (
                <div className="bg-gray-800 border border-gray-700 p-8 rounded-2xl w-full max-w-md shadow-2xl transform transition-all animate-in slide-in-from-bottom-4">
                    <div className="flex items-center space-x-3 mb-6">
                        <CheckCircle className="w-8 h-8 text-green-400" />
                        <h3 className="text-2xl font-bold text-white">Match Found!</h3>
                    </div>
                    <div className="space-y-4">
                        <div>
                            <p className="text-xs text-gray-500 uppercase font-semibold tracking-wider">Title</p>
                            <p className="text-xl text-gray-100 font-medium">{result.title}</p>
                        </div>
                        <div className="w-full h-px bg-gray-700"></div>
                        <div className="flex justify-between items-end">
                            <div>
                                <p className="text-xs text-gray-500 uppercase font-semibold tracking-wider">Confidence</p>
                                <p className="text-blue-400 font-bold">{result.confidence_score}%</p>
                            </div>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
}