interface IKeyValueStore {
	set(key: string, value: any, expireInSeconds: number): void;
	get(key: string): string | null;
	remove(key: string): void;
	clear(): void;
}