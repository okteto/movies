FROM node:16

WORKDIR /src

COPY package.json yarn.lock ./
RUN yarn install

COPY . .
CMD ["yarn", "start"]
