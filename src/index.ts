import { main } from "@blockless/sdk-ts/dist/lib/entry"; // Import directly from submodule
// import { writeOutput } from "@blockless/sdk-ts/dist/lib/stdin"; // This might not be needed?

// Define a type for environment variables
interface EnvVars {
  TARGET?: string;
}

main(async () => {
  try {
    console.log("PhantomDNS Worker starting - initialization check");
    
    // Get environment variables and check if TARGET is empty or undefined
    const env: EnvVars = process.env as any;
    const targetUrl = env.TARGET || "";
    
    console.log(`Environment check - TARGET: ${targetUrl ? "provided" : "missing"}`);
    
    // Show info if no target is specified
    if (!targetUrl || targetUrl.trim() === "") {
      console.log("PhantomDNS Worker is active - No target specified, displaying welcome info");
      
      // Simple plain text response
      const welcomeInfo = 
        "PhantomDNS Worker is active\n" +
        "Worker ID: " + Math.random().toString(36).substring(2, 8) + "\n" +
        "Time: " + new Date().toISOString() + "\n\n" +
        "To use: Add TARGET parameter with the URL to access\n\n" +
        "Status: OK - Worker is running correctly\n" +
        "Permission test: " + await testPermissions();

      return new Response(welcomeInfo, {
        status: 200,
        headers: {
          "Content-Type": "text/plain; charset=utf-8",
          "Cache-Control": "no-store, no-cache",
          "X-Proxy-By": "PhantomDNS",
          "Access-Control-Allow-Origin": "*",
          "Access-Control-Allow-Methods": "GET, POST, OPTIONS",
          "Access-Control-Allow-Headers": "Content-Type"
        }
      });
    }
    
    console.log(`PhantomDNS Worker active - Processing request for: ${targetUrl}`);

    try {
      console.log(`Establishing connection: ${targetUrl}`);
      const response = await fetch(targetUrl, { 
        method: 'GET', 
        headers: { 
          "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
          "Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
          "Accept-Language": "en-US,en;q=0.9",
          "Accept-Encoding": "gzip, deflate, br",
          "sec-ch-ua": "\"Google Chrome\";v=\"120\", \"Chromium\";v=\"120\", \"Not=A?Brand\";v=\"99\"",
          "sec-ch-ua-mobile": "?0",
          "sec-ch-ua-platform": "\"Windows\"",
          "sec-fetch-dest": "document",
          "sec-fetch-mode": "navigate",
          "sec-fetch-site": "none",
          "sec-fetch-user": "?1",
          "upgrade-insecure-requests": "1",
          "priority": "u=0, i",
          "dnt": "1"
        }
      });

      console.log(`Connection status: ${response.status}`);

      // Clone the response to read it multiple times
      const clonedResponse = response.clone();

      if (!response.ok) {
        const errorText = await clonedResponse.text();
        console.error(`Connection error: ${response.status}. Content: ${errorText.substring(0, 100)}`);
        
        // Return simple error text
        const errorInfo = 
          "ERROR: Failed to connect to target\n" +
          "Status: " + response.status + "\n" +
          "Target: " + targetUrl + "\n" +
          "Error details: " + errorText.substring(0, 200);
        
        return new Response(errorInfo, {
          status: response.status,
          headers: {
            "Content-Type": "text/plain; charset=utf-8",
            "X-Proxy-By": "PhantomDNS",
            "Access-Control-Allow-Origin": "*",
            "Cache-Control": "no-store, no-cache"
          }
        });
      }

      // Get content type from original response
      const contentType = response.headers.get("Content-Type") || "text/plain";
      console.log(`Content-Type: ${contentType}`);

      // Return the actual response with all headers preserved
      const responseText = await clonedResponse.text();
      console.log(`Connection successful. Response size: ${responseText.length} bytes`);
      
      // Pass through the original content with minimal modification
      const headers = new Headers();
      
      // Copy all the response headers
      response.headers.forEach((value, key) => {
        headers.set(key, value);
      });
      
      // Add our proxy headers
      headers.set("X-Proxy-By", "PhantomDNS");
      headers.set("Access-Control-Allow-Origin", "*");
      headers.set("Cache-Control", "no-store, no-cache");
      
      return new Response(responseText, {
        status: response.status,
        headers: headers
      });

    } catch (error) {
      console.error(`Error during connection: ${error}`);
      
      // Return simple error text
      const errorInfo = 
        "ERROR: Exception occurred\n" +
        "Message: " + String(error) + "\n" +
        "Target: " + targetUrl + "\n" +
        "This may be a permission issue or Cloudflare protection. Check bls.toml permissions.";
      
      return new Response(errorInfo, {
        status: 500,
        headers: {
          "Content-Type": "text/plain; charset=utf-8",
          "X-Proxy-By": "PhantomDNS",
          "Access-Control-Allow-Origin": "*",
          "Cache-Control": "no-store, no-cache"
        }
      });
    }
  } catch (error) {
    console.error(`Critical worker error: ${error}`);
    return new Response(`Critical Error: ${String(error)}`, {
      status: 500,
      headers: {
        "Content-Type": "text/plain; charset=utf-8",
        "X-Proxy-By": "PhantomDNS",
        "Access-Control-Allow-Origin": "*",
        "Cache-Control": "no-store, no-cache"
      }
    });
  }
});

// Test permissions by making a simple fetch request
async function testPermissions(): Promise<string> {
  try {
    // Try fetching a known working site
    const testResponse = await fetch("https://example.com", {
      method: 'HEAD'
    });
    
    // Check response headers and status
    const headers = Array.from(testResponse.headers.entries())
      .map(([key, value]) => `${key}: ${value}`)
      .join(", ");
    
    return `OK (Status ${testResponse.status}, Headers: ${headers.substring(0, 100)}...)`;
  } catch (error) {
    return `FAILED (${String(error)})`;
  }
} 