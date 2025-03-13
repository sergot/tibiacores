FROM node:lts-alpine

WORKDIR /app

COPY package.json package-lock.json* ./

RUN npm install

RUN npm install -g vite

CMD ["npm", "run", "dev", "--", "--host"]