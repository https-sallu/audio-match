FROM node:18-alpine

WORKDIR /app

COPY frontend/package*.json ./
RUN npm install

COPY frontend/ ./

EXPOSE 80
CMD ["npm", "run", "dev"]