import { getNodeAutoInstrumentations } from '@opentelemetry/auto-instrumentations-node';
import { OTLPTraceExporter } from '@opentelemetry/exporter-trace-otlp-grpc';

import { NodeSDK } from '@opentelemetry/sdk-node';

const { GraphQLInstrumentation } = require('@opentelemetry/instrumentation-graphql');

const sdk = new NodeSDK({
    autoDetectResources: true,
    instrumentations: [getNodeAutoInstrumentations(),
                        new GraphQLInstrumentation({
                            // optional params
                            // allowValues: true,
                            // depth: 2,
                            // mergeItems: true,
                            })],
    traceExporter: new OTLPTraceExporter(),
});

sdk.start();
