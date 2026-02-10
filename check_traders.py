import sqlite3

conn = sqlite3.connect('data/data.db')
cursor = conn.cursor()

print("=== AI Models ===")
cursor.execute('SELECT id, name, provider, enabled, use_browser_automation FROM ai_models')
ai_models = cursor.fetchall()
for model in ai_models:
    print(f"  {model[0]}: {model[1]} - Provider: {model[2]}, Enabled: {model[3]}, BrowserAuto: {model[4]}")

print("\n=== Traders ===")
cursor.execute('SELECT id, name, ai_model, use_browser_automation FROM traders')
traders = cursor.fetchall()
for trader in traders:
    print(f"  {trader[0]}: {trader[1]} - AI Model: {trader[2]}, BrowserAuto: {trader[3]}")

conn.close()