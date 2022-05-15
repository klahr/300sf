FROM node:alpine

RUN mkdir /app
ADD . /app
WORKDIR /app

COPY package*.json ./

RUN npm install
# RUN npm install -g typescript
# RUN npm install -g terser
#
# RUN tsc --build designer/tsconfig-release.json
# RUN terser --compress --mangle -o public/js/designer.js -- public/js/designer.js

EXPOSE 8080

CMD ["node", "server.js"]
