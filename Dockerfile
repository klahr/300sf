FROM node:alpine

RUN mkdir /app
ADD . /app
WORKDIR /app

COPY package*.json ./

RUN npm install
RUN npm install -g typescript
RUN npm install -g terser

RUN tsc --build src/tsconfig-releasee.json
RUN terser --compress --mangle -o public/js/app.js -- public/js/app.js

EXPOSE 3000

CMD ["node", "server.js"]
