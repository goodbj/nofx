/**
 * Table Export Utility for NOFX Trading Dashboard
 * Provides functionality to export tabular data to various formats (currently CSV)
 */

export enum ExportFormat {
  CSV = 'CSV',
  JSON = 'JSON'
}

export interface TableData {
  headers: string[];
  rows: (string | number)[][];
  title?: string;
}

export class TableExporter {
  /**
   * Exports table data to the specified format
   * @param data The table data to export
   * @param filename The name for the output file (without extension)
   * @param format The export format
   */
  static export(data: TableData, filename: string, format: ExportFormat): void {
    switch (format) {
      case ExportFormat.CSV:
        this.exportToCSV(data, filename);
        break;
      case ExportFormat.JSON:
        this.exportToJSON(data, filename);
        break;
      default:
        throw new Error(`Unsupported export format: ${format}`);
    }
  }

  /**
   * Exports table data to CSV format
   * @param data The table data to export
   * @param filename The name for the output file (without extension)
   */
  private static exportToCSV(data: TableData, filename: string): void {
    // Add BOM to support UTF-8 with Chinese characters
    const BOM = '\uFEFF';
    
    // Create header row
    const headerRow = data.headers.join(',');
    
    // Process data rows
    const csvRows = [headerRow];
    
    for (const row of data.rows) {
      // Process each cell in the row
      const processedRow = row.map(cell => {
        // Convert to string and escape commas and quotes
        let cellStr = String(cell);
        
        // If the cell contains commas, newlines, or quotes, wrap in quotes and escape internal quotes
        if (cellStr.includes(',') || cellStr.includes('"') || cellStr.includes('\n')) {
          cellStr = `"${cellStr.replace(/"/g, '""')}"`;
        }
        
        return cellStr;
      });
      
      csvRows.push(processedRow.join(','));
    }
    
    // Create CSV content
    const csvContent = csvRows.join('\n');
    
    // Create blob with BOM
    const blob = new Blob([BOM + csvContent], { type: 'text/csv;charset=utf-8;' });
    
    // Create download link
    const link = document.createElement('a');
    const fullFilename = `${filename}.csv`;
    
    link.setAttribute('href', URL.createObjectURL(blob));
    link.setAttribute('download', fullFilename);
    link.style.visibility = 'hidden';
    
    // Trigger download
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  }

  /**
   * Exports table data to JSON format
   * @param data The table data to export
   * @param filename The name for the output file (without extension)
   */
  private static exportToJSON(data: TableData, filename: string): void {
    // Create JSON structure with headers and rows
    const jsonData = {
      title: data.title || 'Exported Data',
      headers: data.headers,
      rows: data.rows,
      timestamp: new Date().toISOString()
    };
    
    // Convert to JSON string with indentation for readability
    const jsonString = JSON.stringify(jsonData, null, 2);
    
    // Create blob
    const blob = new Blob([jsonString], { type: 'application/json;charset=utf-8;' });
    
    // Create download link
    const link = document.createElement('a');
    const fullFilename = `${filename}.json`;
    
    link.setAttribute('href', URL.createObjectURL(blob));
    link.setAttribute('download', fullFilename);
    link.style.visibility = 'hidden';
    
    // Trigger download
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  }
}