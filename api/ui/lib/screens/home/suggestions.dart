import 'package:flutter/material.dart';

import '../../service/api.dart';

class SuggestionsWidget extends StatefulWidget {
  final ClientAuthentication authentication;

  const SuggestionsWidget({super.key, required this.authentication});

  @override
  State<SuggestionsWidget> createState() => _SuggestionsWidgetState();
}

class _SuggestionsWidgetState extends State<SuggestionsWidget> {
  late Future<Map<String, List<String>>?> _suggestions;

  Future<Map<String, List<String>>?> _loadSuggestions() async {
    Map<String, List<String>>? suggestions;
    String? error;
    (suggestions, error) = await widget.authentication.getSuggestions();
    if (error != null && error.isNotEmpty) {
      _snackbarMessage(error);
    }
    return suggestions;
  }

  void _snackbarMessage(String s) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(s)),
    );
  }

  @override
  void initState() {
    super.initState();
    _suggestions = _loadSuggestions();
  }

  Future<void> _deleteSuggestion(String type, String? value) async {
    await widget.authentication.deleteSuggestion(type, value);
    setState(() {
      _suggestions = _loadSuggestions();
    });
  }

  @override
  Widget build(BuildContext context) {
    return Center(
      child: FutureBuilder<Map<String, List<String>>?>(
        future: _suggestions,
        builder: (context, snapshot) {
          if (snapshot.hasData) {
            Map<String, List<String>>? sugg = snapshot.data;
            if (sugg != null) {
              if (sugg.isEmpty) {
                return const Text(
                    'Currently, there are no suggestions to display.');
              }
              List<Widget> suggList = [];
              for (var s in sugg.entries) {
                for (var v in s.value) {
                  suggList.add(SuggestionWidget(type: s.key, suggestion: v, fnRm: _deleteSuggestion,));
                }
              }
              return SelectionArea(
                  child: SizedBox(
                      width: 500,
                      child: Column(
                        // Column is also a layout widget. It takes a list of children and
                        // arranges them vertically. By default, it sizes itself to fit its
                        // children horizontally, and tries to be as tall as its parent.
                        //
                        // Column has various properties to control how it sizes itself and
                        // how it positions its children. Here we use mainAxisAlignment to
                        // center the children vertically; the main axis here is the vertical
                        // axis because Columns are vertical (the cross axis would be
                        // horizontal).
                        //
                        // TRY THIS: Invoke "debug painting" (choose the "Toggle Debug Paint"
                        // action in the IDE, or press "p" in the console), to see the
                        // wireframe for each widget.
                          mainAxisAlignment: MainAxisAlignment.start,
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: suggList,
                  )));
            }
          }
          return const Scaffold(
              body: Center(child: CircularProgressIndicator()));
        },
      ),
    );
  }
}

class SuggestionWidget extends StatelessWidget {
  final String type;
  final String suggestion;
  final Future<void> Function(String type, String name) fnRm;

  const SuggestionWidget({super.key, required this.type, required this.suggestion, required this.fnRm});

  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.center,
      children: [
        Text('$type >> $suggestion'),
        TextButton(
            onPressed: () => fnRm(type, suggestion),
            child:
            const Icon(Icons.delete_forever_outlined, color: Colors.red)),
      ],
    );
  }
}