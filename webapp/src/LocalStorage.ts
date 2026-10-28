/// <reference path="KeyValueStore.ts"/>

class LocalStorageItem {
	public value: any;
	public expireAt: number;
}

class LocalStorage implements IKeyValueStore {
	public set(key: string, value: any, expireInSeconds: number): void {
		const item = new LocalStorageItem();
		item.value = value;
		if (expireInSeconds > 0) {
			item.expireAt = new Date().getTime() + expireInSeconds * 1000;
		}
		localStorage.setItem(key, JSON.stringify(item));
	}

	public get(key: string): string | null {
		const value = localStorage.getItem(key);
		if (value) {
			const t = JSON.parse(value) as LocalStorageItem;
			if (new Date().getTime() > t.expireAt) {
				localStorage.removeItem(key);
				return null;
			}
			return t.value;
		}
		return null;
	}

	public remove(key: string): void {
		localStorage.removeItem(key);
	}

	public clear(): void {
		localStorage.clear();
	}
}