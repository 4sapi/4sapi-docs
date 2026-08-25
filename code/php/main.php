<?php

// Read credentials from environment variables.
$key = getenv('FOUR_S_API_KEY');
if (!$key) {
    throw new RuntimeException('FOUR_S_API_KEY is required');
}

// Use a default model unless FOUR_S_API_MODEL is provided.
$model = getenv('FOUR_S_API_MODEL') ?: 'gpt-5.6';

// Build the JSON payload for the chat completion request.
$payload = json_encode([
    'model' => $model,
    'messages' => [
        [
            'role' => 'user',
            'content' => 'Introduce 4SAPI in one sentence.',
        ],
    ],
], JSON_UNESCAPED_UNICODE);

// Send the request to the 4SAPI chat completions endpoint.
$ch = curl_init('https://4sapi.com/v1/chat/completions');
curl_setopt_array($ch, [
    CURLOPT_POST => true,
    CURLOPT_POSTFIELDS => $payload,
    CURLOPT_RETURNTRANSFER => true,
    CURLOPT_HTTPHEADER => [
        'Authorization: Bearer ' . $key,
        'Content-Type: application/json',
    ],
]);

$body = curl_exec($ch);
$status = curl_getinfo($ch, CURLINFO_HTTP_CODE);
curl_close($ch);

if ($status >= 300) {
    throw new RuntimeException("HTTP $status: $body");
}

$data = json_decode($body, true);
echo ($data['choices'][0]['message']['content'] ?? $body) . PHP_EOL;
