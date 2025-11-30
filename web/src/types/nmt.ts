export interface NMTHeader {
    name: string;
    reserved: number;
    committed: number;
}

export interface NMTEntry extends NMTHeader { }

export interface NMTReport {
    pid: number;
    created: number;
    nmt_entries: NMTEntry[];
}
