## 2.0.0

- Regenerated with protoc_plugin 25.1 / connect-dart 2.0; requires protobuf ^6.1.0, connectrpc ^2.0.0, Dart ^3.7. WKTs now come from package:protobuf.
- Breaking: `Struct` and `Timestamp` are re-exported from `package:protobuf/well_known_types/...`; the vendored `src/google/protobuf` copies of any, duration, field_mask, struct and timestamp are removed.
- Breaking: the barrel no longer exports the `.pbserver.dart` server stubs.
- Depends on antinvestor_api_common ^2.0.0.

## 1.53.0

- Generated Dart SDK from proto definitions

