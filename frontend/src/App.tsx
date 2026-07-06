import { useState, useEffect } from 'react';
import Recognize from './pages/Recognize';
import { api, Song } from './services/api';
import { Music, Upload, Database, Activity, Trash2, Loader } from 'lucide-react';

export default function App() {
    const [tab, setTab] = useState<'recognize' | 'admin'>('recognize');
    const [file, setFile] = useState<File | null>(null);
    const [isUploading, setIsUploading] = useState(false);
    const [songs, setSongs] = useState<Song[]>([]);
    const [deletingId, setDeletingId] = useState<number | null>(null);

    useEffect(() => {
        if (tab === 'admin') {
            loadSongs();
        }
    }, [tab]);

    const loadSongs = () => {
        // The ?t= timestamp forces the browser to fetch fresh data every single time
        api.getSongs().then(setSongs).catch(console.error);
    };

    const handleUpload = async () => {
        if (!file) return;
        setIsUploading(true);
        try {
            const cleanName = file.name.replace('.wav', '');
            await api.importSong(cleanName, "Unknown Artist", file);
            setFile(null);
            loadSongs();
        } catch (err: any) {
            alert("Upload Failed: " + err.message);
        } finally {
            setIsUploading(false);
        }
    };

    const handleDelete = async (id: number, title: string) => {
        if (!window.confirm(`Are you sure you want to delete "${title}"? This will remove all its fingerprints.`)) {
            return;
        }

        setDeletingId(id);

        try {
            await api.deleteSong(id);
            setSongs((prevSongs) => prevSongs.filter((song) => song.id !== id));
        } catch (err: any) {
            alert("Backend refused to delete: " + err.message);
        } finally {
            setDeletingId(null);
        }
    };

    return (
        <div className="min-h-screen bg-gray-950 text-gray-100 font-sans flex flex-col">
            <header className="border-b border-gray-800 bg-gray-900/50 backdrop-blur-md sticky top-0 z-50">
                <div className="max-w-5xl mx-auto px-6 py-4 flex justify-between items-center">
                    <div className="flex items-center space-x-3 cursor-pointer" onClick={() => setTab('recognize')}>
                        <div className="bg-blue-600 p-2 rounded-lg">
                            <Activity className="w-5 h-5 text-white" />
                        </div>
                        <h1 className="text-2xl font-bold bg-gradient-to-r from-blue-400 to-indigo-400 bg-clip-text text-transparent">
                            AudioMatch
                        </h1>
                    </div>
                    <nav className="flex space-x-2 bg-gray-800 p-1 rounded-lg">
                        <button
                            onClick={() => setTab('recognize')}
                            className={`px-4 py-2 rounded-md font-medium transition-colors ${tab === 'recognize' ? 'bg-gray-700 text-white' : 'text-gray-400 hover:text-gray-200'}`}
                        >
                            Recognize
                        </button>
                        <button
                            onClick={() => setTab('admin')}
                            className={`px-4 py-2 rounded-md font-medium transition-colors flex items-center gap-2 ${tab === 'admin' ? 'bg-gray-700 text-white' : 'text-gray-400 hover:text-gray-200'}`}
                        >
                            <Database className="w-4 h-4" /> Dataset
                        </button>
                    </nav>
                </div>
            </header>

            <main className="flex-grow max-w-5xl mx-auto px-6 py-12 w-full">
                {tab === 'recognize' ? (
                    <Recognize />
                ) : (
                    <div className="space-y-12 animate-in fade-in duration-500">
                        <div className="bg-gray-900 border border-gray-800 rounded-2xl p-8 shadow-xl">
                            <div className="flex items-center space-x-3 mb-6">
                                <Upload className="text-blue-400 w-6 h-6" />
                                <h2 className="text-2xl font-bold text-white">Ingest New Audio</h2>
                            </div>
                            <div className="flex flex-col md:flex-row items-center gap-4">
                                <input
                                    type="file"
                                    accept=".wav"
                                    onChange={(e) => setFile(e.target.files?.[0] || null)}
                                    className="block w-full text-sm text-gray-400 file:mr-4 file:py-3 file:px-6 file:rounded-full file:border-0 file:text-sm file:font-semibold file:bg-gray-800 file:text-blue-400 hover:file:bg-gray-700 cursor-pointer"
                                />
                                <button
                                    // Added flex, items-center, justify-center, and gap-2 to align the new icon perfectly
                                    className={`px-8 py-3 rounded-full font-bold text-white transition-all w-full md:w-auto flex-shrink-0 flex items-center justify-center gap-2 ${isUploading ? 'bg-indigo-600/50 cursor-not-allowed' : 'bg-blue-600 hover:bg-blue-500 shadow-lg shadow-blue-900/20'}`}
                                    onClick={handleUpload}
                                    disabled={isUploading || !file}
                                >
                                    {isUploading ? (
                                        <>
                                            <Loader className="w-5 h-5 animate-spin" />
                                            Processing...
                                        </>
                                    ) : (
                                        "Train Engine"
                                    )}
                                </button>
                            </div>
                        </div>

                        <div>
                            <div className="flex items-center space-x-3 mb-6">
                                <Music className="text-indigo-400 w-6 h-6" />
                                <h2 className="text-2xl font-bold text-white">Learned Audio Signatures</h2>
                            </div>
                            <div className="bg-gray-900 border border-gray-800 rounded-2xl overflow-hidden shadow-xl">
                                <table className="w-full text-left border-collapse">
                                    <thead>
                                        <tr className="bg-gray-800/50 text-gray-400 text-sm uppercase tracking-wider">
                                            <th className="p-5 font-semibold">#</th>
                                            <th className="p-5 font-semibold">Track Title</th>
                                            <th className="p-5 font-semibold">Duration</th>
                                            <th className="p-5 font-semibold">Ingested On</th>
                                            <th className="p-5 font-semibold text-right">Actions</th>
                                        </tr>
                                    </thead>
                                    <tbody className="divide-y divide-gray-800">
                                        {songs.length === 0 ? (
                                            <tr>
                                                <td colSpan={5} className="p-8 text-center text-gray-500 italic">
                                                    Database is currently empty. Train the engine above.
                                                </td>
                                            </tr>
                                        ) : (
                                            songs.map((song, index) => (
                                                <tr key={song.id} className="hover:bg-gray-800/30 transition-colors">
                                                    <td className="p-5 text-gray-500">{index + 1}</td>
                                                    <td className="p-5 font-medium text-gray-200">{song.title}</td>
                                                    <td className="p-5 text-gray-400">{song.duration.toFixed(2)}s</td>
                                                    <td className="p-5 text-gray-500 text-sm">
                                                        {new Date(song.created_at).toLocaleDateString()}
                                                    </td>
                                                    <td className="p-5 text-right">
                                                        <button
                                                            onClick={() => handleDelete(song.id, song.title)}
                                                            disabled={deletingId === song.id}
                                                            className={`p-2 rounded-lg transition-colors ${deletingId === song.id
                                                                ? 'text-blue-400 bg-blue-400/10 cursor-not-allowed'
                                                                : 'text-gray-500 hover:text-red-400 hover:bg-red-400/10'
                                                                }`}
                                                            title="Delete Song"
                                                        >
                                                            {deletingId === song.id ? (
                                                                <Loader className="w-5 h-5 animate-spin" />
                                                            ) : (
                                                                <Trash2 className="w-5 h-5" />
                                                            )}
                                                        </button>
                                                    </td>
                                                </tr>
                                            ))
                                        )}
                                    </tbody>
                                </table>
                            </div>
                        </div>
                    </div>
                )}
            </main>

            <footer className="border-t border-gray-800 bg-gray-900 mt-auto">
                <div className="max-w-5xl mx-auto px-6 py-8 flex flex-col items-center justify-center text-center">
                    <p className="text-gray-400 font-medium">AudioMatch DSP Engine • Crafted by Salman Abbas</p>
                    <p className="text-gray-600 text-sm mt-2">CapregSoft • Wah Cantonment</p>
                </div>
            </footer>
        </div>
    );
}