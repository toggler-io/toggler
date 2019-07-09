# Deploying with docker

## building the img manually

```bash
git clone "https://github.com/adamluzsi/toggler.git"
cd toggler
docker build . -t toggler
docker run --rm -d --network host -e DATABASE_URL=${DATABASE_URL} --name toggler toggler
```
