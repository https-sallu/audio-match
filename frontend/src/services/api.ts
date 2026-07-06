const API_URL = 'http://localhost:8080/api';

export interface MatchResult {
    song_id: number;
    title: string;
    artist: string;
    confidence_score: number;
    total_matches: number;
    offset_consistency: number;
}

export interface Song {
    id: number;
    title: string;
    artist: string;
    duration: number;
    created_at: string;
}

export const api = {
    async importSong(title: string, artist: string, file: File) {
        const formData = new FormData();
        formData.append('title', title);
        formData.append('artist', artist);
        formData.append('audio', file);
        const res = await fetch(`${API_URL}/songs/import`, { method: 'POST', body: formData });
        if (!res.ok) throw new Error('Import failed');
        return res.json();
    },

    async recognizeAudio(wavBlob: Blob): Promise<MatchResult> {
        const formData = new FormData();
        formData.append('audio', wavBlob, 'rec.wav');
        const res = await fetch(`${API_URL}/recognize`, { method: 'POST', body: formData });
        if (res.status === 404) throw new Error('No match found.');
        if (!res.ok) throw new Error('API error.');
        return res.json();
    },

    async getSongs(): Promise<Song[]> {
        // The ?t= timestamp forces the browser to fetch fresh data every single time
        const res = await fetch(`${API_URL}/songs?t=${Date.now()}`, { cache: 'no-store' });
        if (!res.ok) throw new Error('Failed to fetch dataset');
        return res.json();
    },

    // NEW: Delete a song by ID
    async deleteSong(id: number): Promise<void> {
        const res = await fetch(`${API_URL}/songs/${id}`, { method: 'DELETE' });
        if (!res.ok) throw new Error('Failed to delete song');
    }
};