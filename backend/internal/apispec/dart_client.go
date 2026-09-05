package apispec

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
)

// writeDartClient emits one method per operation, so the app cannot call an
// endpoint that no longer exists or pass the wrong body type.
func writeDartClient(b *bytes.Buffer, reg *Registry, ops []Operation) {
	b.WriteString(`/// Typed client for the Carousel Studio API.
///
/// Authentication is a session cookie. Dart's http package does not keep a
/// cookie jar, so the session is captured from Set-Cookie on sign-in and sent
/// back on every later call — see [ApiClient.session].
class ApiClient {
  ApiClient({required this.baseUrl, http.Client? httpClient})
      : _http = httpClient ?? http.Client();

  final String baseUrl;
  final http.Client _http;

  /// The session cookie value, persisted by the caller across launches.
  String? session;

  Uri _uri(String path, [Map<String, String>? query]) {
    final cleaned = <String, String>{
      for (final entry in (query ?? {}).entries)
        if (entry.value.isNotEmpty) entry.key: entry.value,
    };
    return Uri.parse('$baseUrl/api/v1$path')
        .replace(queryParameters: cleaned.isEmpty ? null : cleaned);
  }

  Map<String, String> _headers({bool json = false}) {
    final headers = <String, String>{};
    if (json) headers['Content-Type'] = 'application/json';
    final cookie = session;
    if (cookie != null) headers['Cookie'] = cookie;
    return headers;
  }

  void _captureSession(http.Response res) {
    final raw = res.headers['set-cookie'];
    if (raw == null) return;
    final match = RegExp(r'tks_session=([^;]*)').firstMatch(raw);
    if (match == null) return;
    final value = match.group(1) ?? '';
    session = value.isEmpty ? null : 'tks_session=$value';
  }

  Future<dynamic> _send(String method, String path,
      {Object? body, Map<String, String>? query}) async {
    final request = http.Request(method, _uri(path, query))
      ..headers.addAll(_headers(json: body != null));
    if (body != null) request.body = jsonEncode(body);

    final res = await http.Response.fromStream(await _http.send(request));
    _captureSession(res);

    if (res.statusCode == 204 || res.body.isEmpty) {
      if (res.statusCode >= 300) {
        throw ApiException(res.statusCode, 'unknown', '');
      }
      return null;
    }
    final decoded = jsonDecode(utf8.decode(res.bodyBytes));
    if (res.statusCode >= 300) {
      final error = (decoded is Map<String, dynamic> ? decoded['error'] : null);
      throw ApiException(
        res.statusCode,
        (error is Map ? error['code'] as String? : null) ?? 'unknown',
        (error is Map ? error['message'] as String? : null) ?? '',
      );
    }
    return decoded;
  }

`)

	sorted := append([]Operation(nil), ops...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].OperationID < sorted[j].OperationID })

	for _, op := range sorted {
		writeDartMethod(b, reg, op)
	}

	b.WriteString("  void close() => _http.close();\n}\n")
}

func writeDartMethod(b *bytes.Buffer, reg *Registry, op Operation) {
	params := pathParams(op.Path)

	var args []string
	for _, p := range params {
		args = append(args, "String "+dartField(p))
	}
	if op.Body != nil {
		args = append(args, reg.TypeName(op.Body)+" body")
	}
	var named []string
	var queryNames []string
	for name := range op.Query {
		queryNames = append(queryNames, name)
	}
	sort.Strings(queryNames)
	for _, name := range queryNames {
		named = append(named, "String "+dartField(name)+" = ''")
	}
	if len(named) > 0 {
		args = append(args, "{"+strings.Join(named, ", ")+"}")
	}

	ret := "void"
	if op.Response != nil {
		ret = reg.TypeName(op.Response)
	}

	fmt.Fprintf(b, "  /// %s\n", op.Summary)
	fmt.Fprintf(b, "  Future<%s> %s(%s) async {\n", ret, op.OperationID, strings.Join(args, ", "))

	// Interpolate path params.
	dartPath := op.Path
	for _, p := range params {
		dartPath = strings.ReplaceAll(dartPath, "{"+p+"}", "$"+dartField(p))
	}

	call := fmt.Sprintf("await _send('%s', '%s'", op.Method, dartPath)
	if op.Body != nil {
		call += ", body: body.toJson()"
	}
	if len(queryNames) > 0 {
		var pairs []string
		for _, name := range queryNames {
			pairs = append(pairs, fmt.Sprintf("'%s': %s", name, dartField(name)))
		}
		call += ", query: {" + strings.Join(pairs, ", ") + "}"
	}
	call += ")"

	if op.Response == nil {
		fmt.Fprintf(b, "    %s;\n", call)
	} else {
		fmt.Fprintf(b, "    final data = %s;\n", call)
		fmt.Fprintf(b, "    return %s.fromJson(data as Map<String, dynamic>);\n", ret)
	}
	b.WriteString("  }\n\n")
}
