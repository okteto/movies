#!/bin/bash
set -e

echo "🎬 Movies App - Divert Setup"
echo "============================"
echo ""

RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}Checking prerequisites...${NC}"
command -v okteto >/dev/null 2>&1 || { echo -e "${RED}okteto CLI is required${NC}" >&2; exit 1; }

SHARED_NS="${1:-movies-staging}"
echo -e "${BLUE}Step 1: Creating shared namespace: ${SHARED_NS}${NC}"

okteto preview deploy \
  --repository https://github.com/okteto/movies \
  --branch divert-demo \
  --label okteto-shared \
  --name ${SHARED_NS} \
  --wait

echo -e "${GREEN}✅ Shared namespace ready at: https://movies-${SHARED_NS}.okteto.dev${NC}"

echo -e "${BLUE}Step 2: Creating your diverted namespace...${NC}"
read -p "Enter your name for the namespace (e.g., alice): " USER_NAME

DIVERT_NS="${USER_NAME}-movies"
okteto namespace create ${DIVERT_NS}

echo -e "${GREEN}✅ Created namespace: ${DIVERT_NS}${NC}"

echo -e "${BLUE}Step 3: Choose what to work on:${NC}"
echo "1) Frontend"
echo "2) Catalog service"
echo "3) API service"
echo "4) Rent service"
read -p "Select option (1-4): " OPTION

case $OPTION in
    1)
        MANIFEST="okteto-frontend-divert.yaml"
        ;;
    2)
        MANIFEST="okteto-catalog-divert.yaml"
        ;;
    3)
        MANIFEST="okteto-api-divert.yaml"
        ;;
    4)
        MANIFEST="okteto-rent-divert.yaml"
        ;;
    *)
        echo -e "${RED}Invalid option${NC}"
        exit 1
        ;;
esac

echo -e "${BLUE}Deploying with Divert...${NC}"
okteto deploy \
  --namespace ${DIVERT_NS} \
  --file ${MANIFEST} \
  --var SHARED_NAMESPACE=${SHARED_NS} \
  --wait

echo -e "${GREEN}✅ Deployment complete!${NC}"
echo ""
echo "Access your environment:"
echo -e "${BLUE}Direct URL:${NC} https://movies-${DIVERT_NS}.okteto.dev"
echo -e "${BLUE}With header:${NC} curl -H \"baggage: okteto-divert=${DIVERT_NS}\" https://movies-${SHARED_NS}.okteto.dev"
