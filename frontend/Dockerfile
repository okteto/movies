FROM node:14 as dev

WORKDIR /src

COPY package.json yarn.lock ./
RUN yarn install
COPY . .

RUN yarn build

FROM nginx:alpine
COPY --from=dev /src/dist /usr/share/nginx/html
EXPOSE 80
