import {TextField, TextFieldProps} from '@mui/material';
import React from 'react';

export interface NumberFieldProps {
    value: number;
    min: number;
    onChange: (value: number) => void;
}

export const NumberField = ({
    value,
    min,
    onChange,
    ...props
}: NumberFieldProps & Omit<TextFieldProps, 'value' | 'onChange'>) => {
    const [stringValue, setStringValue] = React.useState<string>(value.toString());
    const [error, setError] = React.useState('');

    return (
        <TextField
            value={stringValue}
            type="number"
            helperText={error}
            error={error !== ''}
            onChange={(event) => {
                setStringValue(event.target.value);
                const i = parseInt(event.target.value, 10);
                if (Number.isNaN(i)) {
                    setError('Invalid number');
                    return;
                }

                if (i < min) {
                    setError('Number must be at least ' + min);
                    return;
                }
                onChange(i);
                setError('');
            }}
            {...props}
        />
    );
};
