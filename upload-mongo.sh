#!/bin/bash

# Verifica se a variável MONGO_URI está definida
if [ -z "$MONGO_URI" ]; then
    echo "Erro: A variável de ambiente MONGO_URI não está definida."
    exit 1
fi

# Configuração do banco e coleção
DATABASE="festa_db"
COLLECTION="guests"
CSV_FILE="guests.csv"

# Verifica se o arquivo CSV existe
if [ ! -f "$CSV_FILE" ]; then
    echo "Erro: Arquivo CSV não encontrado!"
    exit 1
fi

# Executa mongoimport dentro de um contêiner Docker
docker run --rm -v "$(pwd):/data" mongo:latest mongoimport \
  --uri "$MONGO_URI" \
  --db "$DATABASE" \
  --collection "$COLLECTION" \
  --type csv \
  --headerline \
  --file /data/"$CSV_FILE"

echo "✅ Importação concluída com sucesso!"
