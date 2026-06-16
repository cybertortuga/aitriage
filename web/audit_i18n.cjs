const fs = require('fs');
const path = require('path');

const srcDir = path.join(__dirname, 'src');
const localesDir = path.join(srcDir, 'locales', 'ru');

// 1. Load all Russian translations
const namespaces = {};
if (fs.existsSync(localesDir)) {
  fs.readdirSync(localesDir).forEach(file => {
    if (file.endsWith('.json')) {
      const ns = file.replace('.json', '');
      try {
        namespaces[ns] = JSON.parse(fs.readFileSync(path.join(localesDir, file), 'utf8'));
      } catch (e) {
        console.error(`Error parsing ${file}: ${e.message}`);
      }
    }
  });
}

function keyExists(obj, keyPath) {
  const parts = keyPath.split('.');
  let current = obj;
  for (const part of parts) {
    if (current === undefined || current === null) return false;
    current = current[part];
  }
  return current !== undefined;
}

// 2. Find all t('...') or t("...") calls in .tsx and .ts files
const missingKeys = new Set();
const foundKeys = new Set();

function walk(dir) {
  const files = fs.readdirSync(dir);
  for (const file of files) {
    const fullPath = path.join(dir, file);
    if (fs.statSync(fullPath).isDirectory()) {
      walk(fullPath);
    } else if (fullPath.endsWith('.tsx') || fullPath.endsWith('.ts')) {
      const content = fs.readFileSync(fullPath, 'utf8');
      
      // Look for t('namespace:key') or t('key')
      // Also handles t("key") and t(\`key\`)
      const regex = /t\(['"\`]?([a-zA-Z0-9_\.:\-]+)['"\`]?/g;
      let match;
      while ((match = regex.exec(content)) !== null) {
        const fullKey = match[1];
        if (!fullKey) continue;
        
        let ns = 'translation'; // default
        let key = fullKey;
        
        if (fullKey.includes(':')) {
          [ns, key] = fullKey.split(':');
        } else {
          // If no namespace provided, usually it comes from useTranslation('ns')
          // Try to guess from the file's useTranslation
          const useTransMatch = /useTranslation\(['"\`]?([a-zA-Z0-9_]+)['"\`]?\)/.exec(content);
          if (useTransMatch && useTransMatch[1]) {
            ns = useTransMatch[1];
          }
        }
        
        if (!namespaces[ns] || !keyExists(namespaces[ns], key)) {
          missingKeys.add(`Missing: ${ns}:${key} (found in ${path.relative(__dirname, fullPath)})`);
        } else {
          foundKeys.add(`${ns}:${key}`);
        }
      }
      
      // 3. Very rudimentary check for hardcoded English in JSX
      // Looking for >[A-Z][a-z]+< or similar
      const hardcodedRegex = />\s*([A-Z][a-zA-Z\s]+)\s*<\/[a-zA-Z0-9]+>/g;
      let hcMatch;
      while ((hcMatch = hardcodedRegex.exec(content)) !== null) {
        const text = hcMatch[1].trim();
        if (text.length > 2 && /^[A-Za-z\s]+$/.test(text)) {
          // Ignore simple common short stuff or variables if accidentally matched
          if (!['div', 'span', 'p', 'h1', 'h2', 'h3', 'br', 'td', 'th'].includes(text.toLowerCase())) {
            // console.log(`Potential hardcoded: "${text}" in ${path.relative(__dirname, fullPath)}`);
          }
        }
      }
    }
  }
}

walk(srcDir);

console.log('=== AUDIT RESULTS ===');
console.log(`Found ${missingKeys.size} missing translation keys.`);
missingKeys.forEach(k => console.log(k));
console.log(`Found ${foundKeys.size} valid translation keys in use.`);
