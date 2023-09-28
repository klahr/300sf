namespace Utils {
    export function isValidDate(date: Date): boolean {
        const d = new Date(date);
        return d.getFullYear() !== 1 && d.getDate() !== 1
    }
}