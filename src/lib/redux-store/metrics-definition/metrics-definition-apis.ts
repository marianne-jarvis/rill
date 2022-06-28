import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import {
  addManyMetricsDefs,
  addOneMetricsDef,
  removeMetricsDef,
  updateMetricsDef,
} from "$lib/redux-store/metrics-definition/metrics-definition-slice";
import { generateApis } from "$lib/redux-store/utils/api-utils";
import * as reduxToolkit from "@reduxjs/toolkit";
import { streamingFetchWrapper } from "$lib/util/fetchWrapper";
import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import type { DimensionDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
import {
  addOneMeasure,
  clearMeasuresForMetricsDefId,
} from "$lib/redux-store/measure-definition/measure-definition-slice";
import {
  addOneDimension,
  clearDimensionsForMetricsDefId,
} from "$lib/redux-store/dimension-definition/dimension-definition-slice";
import { asyncWait } from "$common/utils/waitUtils";

const { createAsyncThunk } = reduxToolkit;

export const {
  fetchManyApi: fetchManyMetricsDefsApi,
  createApi: createMetricsDefsApi,
  updateApi: updateMetricsDefsApi,
  deleteApi: deleteMetricsDefsApi,
} = generateApis<EntityType.MetricsDefinition>(
  [EntityType.MetricsDefinition, "metricsDefinition", "metrics"],
  [addManyMetricsDefs, addOneMetricsDef, updateMetricsDef, removeMetricsDef],
  []
);

export const generateMeasuresAndDimensionsApi = createAsyncThunk(
  `${EntityType.MetricsDefinition}/generateMeasuresAndDimensions`,
  async (id: string, thunkAPI) => {
    const stream = streamingFetchWrapper<
      MeasureDefinitionEntity | DimensionDefinitionEntity
    >(`metrics/${id}/generate-measures-dimensions`, "POST");
    thunkAPI.dispatch(clearMeasuresForMetricsDefId(id));
    thunkAPI.dispatch(clearDimensionsForMetricsDefId(id));
    await asyncWait(10);
    for await (const measureOrDimension of stream) {
      console.log(measureOrDimension);
      if (measureOrDimension.type === EntityType.MeasureDefinition) {
        thunkAPI.dispatch(
          addOneMeasure(measureOrDimension as MeasureDefinitionEntity)
        );
      } else {
        thunkAPI.dispatch(
          addOneDimension(measureOrDimension as DimensionDefinitionEntity)
        );
      }
    }
  }
);
