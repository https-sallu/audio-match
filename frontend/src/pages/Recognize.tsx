import { useState, useEffect } from 'react';
import { useAudioRecorder } from '../hooks/useAudioRecorder';
import { api, MatchResult } from '../services/api';
import { Mic, Loader, CheckCircle, Search } from 'lucide-react';

export default function Recognize() {
    const { isRecording, startRecording, recordedWav, error: micErr } = useAudioRecorder();
    const [isProcessing, setIsProcessing] = useState(false);
    const [result, setResult] = useState<MatchResult | null>(null);
    const [error, setError] = useState<string | null>(null);

    // Clear previous results the moment we start recording
    const handleRecordClick = () => {
        setResult(null);
        setError(null);
        startRecording();
    };

    useEffect(() => {
        if (recordedWav) {
            setIsProcessing(true);
            api.recognizeAudio(recordedWav)
                .then(setResult)
                .catch(e => setError(e.message))
                .finally(() => setIsProcessing(false));
        }
    }, [recordedWav]);

    return (
        <div className="flex flex-col items-center justify-center min-h-[60vh] space-y-12">
            <div className="text-center space-y-4">
                <h2 className="text-4xl font-extrabold text-white tracking-tight">Tap to Shazam</h2>
                <p className="text-gray-400">Discover the music playing around you</p>
            </div>

            <div className="relative group flex flex-col items-center">
                {/* Radar Ripple Animation Background */}
                {isRecording && (
                    <div className="absolute inset-0 bg-blue-500 rounded-full animate-ping opacity-75 scale-150"></div>
                )}

                <button
                    onClick={handleRecordClick}
                    disabled={isRecording || isProcessing}
                    className={`relative z-10 p-10 rounded-full flex items-center justify-center shadow-2xl transition-all duration-300 ${isRecording
                        ? 'bg-red-500 scale-110 shadow-red-500/50'
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
                    {isRecording && <p className="text-blue-400 font-medium animate-pulse flex items-center gap-2"><Mic className="w-4 h-4" /> Listening to audio...</p>}
                    {isProcessing && <p className="text-yellow-400 font-medium animate-pulse flex items-center gap-2"><Search className="w-4 h-4" /> Analyzing frequencies...</p>}
                </div>
            </div>

            {(micErr || error) && !isRecording && !isProcessing && (
                <div className="text-red-400 bg-red-900/20 px-6 py-4 rounded-xl border border-red-900/50 shadow-lg">
                    {micErr || error}
                </div>
            )}

            {result && !isRecording && !isProcessing && (
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