FROM node:20-alpine

WORKDIR /app

# Copy package.json and package-lock.json
COPY package.json package-lock.json* ./

# Install dependencies
RUN npm install

RUN npm install -g next

# Expose the port
EXPOSE 3000

# Start the development server
CMD ["npm", "run", "dev"] 