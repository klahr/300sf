namespace Responses {
    export class ListAvailable {
        public list: Models.List;
        public available: boolean;
    }
    export class ListsSpecies {
        public result: Array<ListAvailable>;
    }
}