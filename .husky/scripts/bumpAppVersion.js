import fs from 'fs'
import path from 'path'
import { fileURLToPath } from 'url'

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)

const packagePath = path.join(__dirname, '..', '..', 'package.json')

const packageJson = JSON.parse(fs.readFileSync(packagePath, 'utf8'))

const versionParts = packageJson.version
  .split('.')
  .map((part) => parseInt(part, 10))

versionParts[2] += 1

packageJson.version = versionParts.join('.')
fs.writeFileSync(packagePath, JSON.stringify(packageJson, null, 2) + '\n')
