import 'dart:io';
import 'dart:convert';

import 'package:flutter/foundation.dart';
import 'package:http/http.dart' as http;

class BaseCrud {
  static String baseUrl() {
    if (kDebugMode) {
      return 'http://localhost:8080';
    }
    return '';
  }
}

class ClientAuthentication extends BaseCrud {
  String? userId;
  String? token;

  Future<(String? error,)> generateVerificationCode(String userId) async {
    this.userId = userId;
    String url = '${BaseCrud.baseUrl()}/api/token/verification/$userId';
    final response = await http.post(Uri.parse(url), headers: <String, String>{
      HttpHeaders.contentTypeHeader: 'application/json; charset=UTF-8',
    });
    if (response.statusCode == 204) { return (null,);} // OPTIONS
    if (response.statusCode == 201) {
      return (null,);
    }
    Map<String, dynamic> responseMap = jsonDecode(response.body);
    return ('${responseMap['error']} (${response.statusCode})',);
  }

  Future<(String? token, String? errorMsg)> validateVerificationCode(
      String verification) async {
    String url = '${BaseCrud.baseUrl()}/api/token/grant/$userId/$verification';
    final response = await http.post(Uri.parse(url), headers: <String, String>{
      HttpHeaders.contentTypeHeader: 'application/json; charset=UTF-8',
    });
    if (response.statusCode == 204) { return (null, null);} // OPTIONS
    Map<String, dynamic> responseMap = jsonDecode(response.body);
    if (response.statusCode == 200) {
      token = responseMap['token'];
      return (token, null);
    }
    return (null, '${responseMap['error']} (${response.statusCode})');
  }

  Future<int> revokeAccess() async {
    String url = '${BaseCrud.baseUrl()}/api/token/revoke/$token';
    final response = await http.post(Uri.parse(url), headers: <String, String>{
      HttpHeaders.contentTypeHeader: 'application/json; charset=UTF-8',
    });
    Map<String, dynamic> responseMap = jsonDecode(response.body);
    int affected = responseMap['affected'];
    return affected;
  }
}
