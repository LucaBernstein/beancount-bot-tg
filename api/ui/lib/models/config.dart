enum ConfigProperty {
  EnableApi(name: 'user.enableApi'),
  IsAdmin(name: 'user.isAdmin'),
  Currency(name: 'user.currency'),
  VacationTag(name: 'user.vacationTag'),
  TimezoneOffset(name: 'user.tzOffset'),
  OmitLeadingSlash(name: 'user.omitCommandSlash');

  const ConfigProperty({required this.name});

  final String name;
}

class Config {
  final bool enableApi;
  final bool isAdmin;
  String? currency;
  String? vacationTag;
  int timezoneOffset;
  bool omitLeadingCommandSlash;

  Config(this.enableApi, this.isAdmin, this.currency, this.vacationTag,
      this.timezoneOffset, this.omitLeadingCommandSlash);

  Map<String, dynamic> diffChanged(Config cnf) {
    Map<String, dynamic> diff = {};
    if (currency != cnf.currency) {
      diff[ConfigProperty.Currency.name] = cnf.currency;
    }
    if (vacationTag != cnf.vacationTag) {
      diff[ConfigProperty.VacationTag.name] = cnf.vacationTag;
    }
    if (timezoneOffset != cnf.timezoneOffset) {
      diff[ConfigProperty.TimezoneOffset.name] = cnf.timezoneOffset;
    }
    if (omitLeadingCommandSlash != cnf.omitLeadingCommandSlash) {
      diff[ConfigProperty.OmitLeadingSlash.name] = cnf.omitLeadingCommandSlash;
    }
    return diff;
  }
}
