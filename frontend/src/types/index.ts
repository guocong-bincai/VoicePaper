export interface Article {
    id: number;
    title: string;
    content: string;
    audio_path: string;
    status: 'pending' | 'processing' | 'completed' | 'failed';
    sentences?: Sentence[];
}

export interface Sentence {
    id: number;
    text: string;
    start_time: number;
    end_time: number;
    order: number;
}
