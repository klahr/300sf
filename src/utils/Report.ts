namespace Utils {
    export function numberOfUniqueSpeciesInReports(reports: Array<Models.Report>): number {
        const uniqueUuids = {};
        reports.forEach(obj => {
            uniqueUuids[obj.species.uuid] = true;
        });
        return Object.keys(uniqueUuids).length;
    }
}